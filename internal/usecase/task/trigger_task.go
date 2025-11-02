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
}

// NewTriggerTaskUseCase 创建触发任务用例
func NewTriggerTaskUseCase(
	taskRepo repository.TaskRepository,
	taskDetailRepo repository.TaskDetailRepository,
	ruleEngine output.RuleEngine,
	observerRegistry output.TaskObserverRegistry,
	distributedLock output.DistributedLock,
) *TriggerTaskUseCase {
	return &TriggerTaskUseCase{
		taskRepo:         taskRepo,
		taskDetailRepo:   taskDetailRepo,
		ruleEngine:       ruleEngine,
		observerRegistry: observerRegistry,
		distributedLock:  distributedLock,
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
	for _, task := range validTasks {
		if err := uc.processTask(ctx, task, expressFuncs, expressArgs, input.TaskMode.GetUniqueFlag()); err != nil {
			fmt.Printf("[TriggerTask] Process task %d failed: %v\n", task.ID, err)
			continue
		}
	}

	return nil
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

	// 通知观察者
	if err := uc.observerRegistry.Notify(ctx, detail); err != nil {
		fmt.Printf("[TriggerTask] Notify observers failed: %v\n", err)
		// 继续执行，不中断流程
	}

	return nil
}

