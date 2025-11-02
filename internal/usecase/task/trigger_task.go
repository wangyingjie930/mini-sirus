package task

import (
	"context"
	"errors"
	"fmt"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/repository"
	"mini-sirus/internal/domain/valueobject"
	"mini-sirus/internal/usecase/dto"
	"mini-sirus/internal/usecase/port/output"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

// TriggerTaskUseCase 触发任务用例
type TriggerTaskUseCase struct {
	taskRepo         repository.TaskRepository
	taskDetailRepo   repository.TaskDetailRepository
	ruleEngine       output.RuleEngine
	observerRegistry output.TaskObserverRegistry
	distributedLock  output.DistributedLock
	riskCheckService output.RiskCheckService // 风控服务应该作为依赖注入，而不是观察者
}

// NewTriggerTaskUseCase 创建触发任务用例
func NewTriggerTaskUseCase(
	taskRepo repository.TaskRepository,
	taskDetailRepo repository.TaskDetailRepository,
	ruleEngine output.RuleEngine,
	observerRegistry output.TaskObserverRegistry,
	distributedLock output.DistributedLock,
	riskCheckService output.RiskCheckService,
) *TriggerTaskUseCase {
	return &TriggerTaskUseCase{
		taskRepo:         taskRepo,
		taskDetailRepo:   taskDetailRepo,
		ruleEngine:       ruleEngine,
		observerRegistry: observerRegistry,
		distributedLock:  distributedLock,
		riskCheckService: riskCheckService,
	}
}

// Execute 执行触发任务用例
func (uc *TriggerTaskUseCase) Execute(ctx context.Context, input dto.TriggerTaskInput) error {
	if input.TaskMode == nil {
		return errors.New("task mode is required")
	}

	userID := input.TaskMode.GetUserID()
	taskType := input.TaskMode.GetTaskType()

	// 用户粒度任务锁
	lockKey := fmt.Sprintf("task_lock:%d:%s", userID, taskType)
	lockID, err := uc.distributedLock.Lock(ctx, lockKey, 30) // 30秒超时
	if err != nil {
		return fmt.Errorf("acquire lock failed: %w", err)
	}
	defer uc.distributedLock.Unlock(ctx, lockKey, lockID)

	fmt.Printf("[TriggerTask] Processing task for user: %d, type: %s\n", userID, taskType)

	// 获取用户待处理任务
	tasks, err := uc.taskRepo.ListByUserIDAndType(ctx, userID, taskType)
	if err != nil {
		return fmt.Errorf("list user tasks failed: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Printf("[TriggerTask] No pending tasks for user: %d\n", userID)
		return nil
	}

	// 过滤有效任务
	validTasks := uc.filterValidTasks(tasks)

	// 获取表达式参数和函数
	expressArgs := input.TaskMode.GetExpressionArguments()
	expressFuncs := uc.buildExpressionFunctions(input.TaskMode)

	// 任务达成判定
	var lastError error
	for _, task := range validTasks {
		if err := uc.processTask(ctx, task, expressFuncs, expressArgs, input.TaskMode.GetUniqueFlag()); err != nil {
			fmt.Printf("[TriggerTask] Process task %d failed: %v\n", task.ID, err)
			lastError = err
			// 如果是风控检查失败，立即返回错误，不继续处理后续任务
			if isRiskCheckError(err) {
				return err
			}
			continue
		}
	}

	return lastError
}

// filterValidTasks 过滤有效的任务
func (uc *TriggerTaskUseCase) filterValidTasks(tasks []*entity.ActUserTask) []*entity.ActUserTask {
	validTasks := make([]*entity.ActUserTask, 0, len(tasks))

	for _, task := range tasks {
		// 过滤已完成的任务
		if task.IsCompleted() {
			continue
		}

		// 检查任务是否过期（30天）
		if task.IsExpired(30) {
			continue
		}

		validTasks = append(validTasks, task)
	}

	return validTasks
}

// buildExpressionFunctions 构建表达式函数
func (uc *TriggerTaskUseCase) buildExpressionFunctions(taskMode dto.TaskModeDTO) map[string]govaluate.ExpressionFunction {
	functions := make(map[string]govaluate.ExpressionFunction)

	// 注册通用函数
	functions["WITH_ANY_TOPIC"] = uc.withAnyTopicFunc()
	functions["LIKE_COUNT_GTE"] = uc.likeCountGteFunc()
	functions["IS_AUDITED"] = uc.isAuditedFunc()
	functions["IS_TODAY"] = uc.isTodayFunc()

	return functions
}

// withAnyTopicFunc 判断是否包含任意话题
func (uc *TriggerTaskUseCase) withAnyTopicFunc() govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return false, errors.New("WITH_ANY_TOPIC requires 2 arguments")
		}

		carryIDs, ok := args[0].([]uint64)
		if !ok {
			return false, errors.New("first argument must be []uint64")
		}

		condIDs, ok := args[1].([]uint64)
		if !ok {
			return false, errors.New("second argument must be []uint64")
		}

		for _, cid := range carryIDs {
			for _, tid := range condIDs {
				if cid == tid {
					return true, nil
				}
			}
		}
		return false, nil
	}
}

// likeCountGteFunc 判断点赞数是否达标
func (uc *TriggerTaskUseCase) likeCountGteFunc() govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return false, errors.New("LIKE_COUNT_GTE requires 2 arguments")
		}

		likeCount, ok := args[0].(float64)
		if !ok {
			return false, errors.New("first argument must be number")
		}

		minCount, ok := args[1].(float64)
		if !ok {
			return false, errors.New("second argument must be number")
		}

		return likeCount >= minCount, nil
	}
}

// isAuditedFunc 判断是否已审核通过
func (uc *TriggerTaskUseCase) isAuditedFunc() govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return false, errors.New("IS_AUDITED requires 1 argument")
		}

		isAudited, ok := args[0].(bool)
		if !ok {
			return false, errors.New("argument must be bool")
		}

		return isAudited, nil
	}
}

// isTodayFunc 判断是否今天
func (uc *TriggerTaskUseCase) isTodayFunc() govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		return true, nil
	}
}

// processTask 处理单个任务
func (uc *TriggerTaskUseCase) processTask(
	ctx context.Context,
	task *entity.ActUserTask,
	functions map[string]govaluate.ExpressionFunction,
	args valueobject.ExpressionArguments,
	uniqueFlag string,
) error {
	// 执行规则引擎判定
	reach, err := uc.ruleEngine.Evaluate(ctx, task.TaskCondExpr, functions, args)
	if err != nil {
		return fmt.Errorf("evaluate expression failed: %w", err)
	}

	if !reach {
		fmt.Printf("[TriggerTask] Task %d not reached\n", task.ID)
		return nil
	}

	fmt.Printf("[TriggerTask] Task %d reached!\n", task.ID)

	// ========== 风控检查（同步执行，阻塞任务完成）==========
	if err := uc.performRiskCheck(ctx, task.UserID, task.ID); err != nil {
		fmt.Printf("[TriggerTask] Risk check failed for user %d: %v\n", task.UserID, err)
		return fmt.Errorf("风控检查失败: %w", err)
	}
	fmt.Printf("[TriggerTask] Risk check passed for user %d\n", task.UserID)

	// 创建任务明细
	detail := &entity.ActUserTaskDetail{
		TaskID:      task.ID,
		UserID:      task.UserID,
		Status:      entity.TaskDetailStatusDone,
		UniqueFlag:  uniqueFlag,
		RewardValue: 1, // 根据实际业务配置
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存任务明细
	if err := uc.taskDetailRepo.Create(ctx, detail); err != nil {
		return fmt.Errorf("save task detail failed: %w", err)
	}

	// 更新任务进度
	task.UpdateProgress()
	if err := uc.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task progress failed: %w", err)
	}

	// 记录任务完成事件（用于风控统计）
	if err := uc.riskCheckService.RecordTaskCompletion(ctx, task.UserID, task.ID, detail.CreatedAt); err != nil {
		fmt.Printf("[TriggerTask] Record task completion failed: %v\n", err)
		// 记录失败不影响任务完成
	}

	// 通知观察者（触达服务、统计服务等非阻塞操作）
	if err := uc.observerRegistry.Notify(ctx, detail); err != nil {
		fmt.Printf("[TriggerTask] Notify observers failed: %v\n", err)
		// 继续执行，不中断流程
	}

	return nil
}

// performRiskCheck 执行风控检查（同步阻塞）
func (uc *TriggerTaskUseCase) performRiskCheck(ctx context.Context, userID, taskID int64) error {
	// 1. 检查用户是否在黑名单中
	isBlacklisted, err := uc.riskCheckService.IsUserBlacklisted(ctx, userID)
	if err != nil {
		return fmt.Errorf("检查黑名单失败: %w", err)
	}
	if isBlacklisted {
		return fmt.Errorf("用户已被列入黑名单，禁止完成任务")
	}

	// 2. 检查用户行为异常
	// 注意：这里传nil作为detail，因为任务还未完成
	if err := uc.riskCheckService.CheckUserBehavior(ctx, userID, nil); err != nil {
		fmt.Printf("[RiskCheck] 用户行为检查失败: %v\n", err)
		// 加入黑名单
		_ = uc.riskCheckService.AddToBlacklist(ctx, userID, "用户行为异常")
		return err
	}

	// 3. 检查任务完成频率
	if err := uc.riskCheckService.CheckTaskFrequency(ctx, userID, taskID); err != nil {
		fmt.Printf("[RiskCheck] 任务频率检查失败: %v\n", err)
		// 频率过高也加入黑名单
		_ = uc.riskCheckService.AddToBlacklist(ctx, userID, "任务完成频率过高")
		return err
	}

	// 4. 检查设备指纹（简化版）
	// 注意：这里传nil作为detail，因为任务还未完成
	if err := uc.riskCheckService.CheckDeviceFingerprint(ctx, userID, nil); err != nil {
		fmt.Printf("[RiskCheck] 设备指纹检查失败: %v\n", err)
		// 设备异常也加入黑名单
		_ = uc.riskCheckService.AddToBlacklist(ctx, userID, "设备指纹异常")
		return err
	}

	return nil
}

// isRiskCheckError 判断是否为风控相关的错误
func isRiskCheckError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 检查是否包含风控相关的关键词
	return strings.Contains(errMsg, "风控检查失败") ||
		strings.Contains(errMsg, "黑名单") ||
		strings.Contains(errMsg, "用户行为异常") ||
		strings.Contains(errMsg, "任务完成频率过高") ||
		strings.Contains(errMsg, "设备指纹异常")
}

