package observer

import (
	"context"
	"fmt"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/usecase/port/output"
)

// CheckinReachObserver 签到触达观察者
type CheckinReachObserver struct {
	reachService output.ReachService
}

// NewCheckinReachObserver 创建签到触达观察者
func NewCheckinReachObserver(reachService output.ReachService) *CheckinReachObserver {
	return &CheckinReachObserver{
		reachService: reachService,
	}
}

// OnTaskDetailCreated 当任务明细创建时
func (o *CheckinReachObserver) OnTaskDetailCreated(ctx context.Context, detail *entity.ActUserTaskDetail) error {
	// 只处理已完成的任务
	if !detail.IsCompleted() {
		return nil
	}

	// 发送触达消息
	params := map[string]interface{}{
		"task_id":      detail.TaskID,
		"reward_value": detail.RewardValue,
	}

	return o.reachService.Send(ctx, "act_checkin_task_detail_done", detail.UserID, params)
}

// OnTaskCompleted 当任务完成时
func (o *CheckinReachObserver) OnTaskCompleted(ctx context.Context, task *entity.ActUserTask) error {
	// 签到任务完成逻辑
	fmt.Printf("[CheckinReachObserver] Task %d completed for user %d\n", task.ID, task.UserID)
	return nil
}

// GetObserverName 获取观察者名称
func (o *CheckinReachObserver) GetObserverName() string {
	return "checkin_reach_observer"
}

// RiskCheckObserver 风控检查观察者
type RiskCheckObserver struct {
	riskCheckService output.RiskCheckService
}

// NewRiskCheckObserver 创建风控检查观察者
func NewRiskCheckObserver(riskCheckService output.RiskCheckService) *RiskCheckObserver {
	return &RiskCheckObserver{
		riskCheckService: riskCheckService,
	}
}

// OnTaskDetailCreated 当任务明细创建时
func (o *RiskCheckObserver) OnTaskDetailCreated(ctx context.Context, detail *entity.ActUserTaskDetail) error {
	fmt.Printf("[RiskCheck] Checking task detail: %d for user: %d\n", detail.ID, detail.UserID)

	// 只对已完成的任务进行风控检查
	if !detail.IsCompleted() {
		return nil
	}

	// 1. 检查用户是否在黑名单中
	isBlacklisted, err := o.riskCheckService.IsUserBlacklisted(ctx, detail.UserID)
	if err != nil {
		return fmt.Errorf("检查黑名单失败: %w", err)
	}
	if isBlacklisted {
		return fmt.Errorf("用户已被列入黑名单，禁止完成任务")
	}

	// 2. 检查用户行为异常
	if err := o.riskCheckService.CheckUserBehavior(ctx, detail.UserID, detail); err != nil {
		fmt.Printf("[RiskCheck] 用户行为检查失败: %v\n", err)
		// 可以选择直接返回错误，或者记录日志后继续
		// 这里选择记录日志并加入黑名单
		_ = o.riskCheckService.AddToBlacklist(ctx, detail.UserID, "用户行为异常")
		return err
	}

	// 3. 检查任务完成频率
	if err := o.riskCheckService.CheckTaskFrequency(ctx, detail.UserID, detail.TaskID); err != nil {
		fmt.Printf("[RiskCheck] 任务频率检查失败: %v\n", err)
		// 频率过高也加入黑名单
		_ = o.riskCheckService.AddToBlacklist(ctx, detail.UserID, "任务完成频率过高")
		return err
	}

	// 4. 检查设备指纹
	if err := o.riskCheckService.CheckDeviceFingerprint(ctx, detail.UserID, detail); err != nil {
		fmt.Printf("[RiskCheck] 设备指纹检查失败: %v\n", err)
		// 设备异常也加入黑名单
		_ = o.riskCheckService.AddToBlacklist(ctx, detail.UserID, "设备指纹异常")
		return err
	}

	// 5. 记录任务完成事件（用于后续的频率统计）
	if err := o.riskCheckService.RecordTaskCompletion(ctx, detail.UserID, detail.TaskID, detail.CreatedAt); err != nil {
		fmt.Printf("[RiskCheck] 记录任务完成事件失败: %v\n", err)
		// 记录失败不影响任务完成
	}

	fmt.Printf("[RiskCheck] 风控检查通过 - User: %d, Task: %d\n", detail.UserID, detail.TaskID)
	return nil
}

// OnTaskCompleted 当任务完成时
func (o *RiskCheckObserver) OnTaskCompleted(ctx context.Context, task *entity.ActUserTask) error {
	fmt.Printf("[RiskCheck] Task %d completed check for user %d\n", task.ID, task.UserID)
	return nil
}

// GetObserverName 获取观察者名称
func (o *RiskCheckObserver) GetObserverName() string {
	return "risk_check_observer"
}

