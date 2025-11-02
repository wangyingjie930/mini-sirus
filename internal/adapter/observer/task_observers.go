package observer

import (
	"context"
	"fmt"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/usecase/port/output"
)

// CheckinReachObserver 签到触达观察者
// 观察者模式适用于：触达通知、异步统计、日志记录等非阻塞操作
// 不适用于：风控检查、权限验证等需要阻塞业务流程的操作
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


