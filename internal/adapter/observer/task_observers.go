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
type RiskCheckObserver struct{}

// NewRiskCheckObserver 创建风控检查观察者
func NewRiskCheckObserver() *RiskCheckObserver {
	return &RiskCheckObserver{}
}

// OnTaskDetailCreated 当任务明细创建时
func (o *RiskCheckObserver) OnTaskDetailCreated(ctx context.Context, detail *entity.ActUserTaskDetail) error {
	// 执行风控检查逻辑
	fmt.Printf("[RiskCheck] Checking task detail: %d for user: %d\n", detail.ID, detail.UserID)
	// TODO: 实际的风控逻辑
	// 1. 检查用户行为异常
	// 2. 检查任务完成频率
	// 3. 检查设备指纹
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

