package notification

import (
	"context"
	"fmt"
	"mini-sirus/internal/domain/entity"
)

// ReachAdapter 触达服务适配器
type ReachAdapter struct {
	// 可以添加实际的触达服务客户端
	// client ReachClient
}

// NewReachAdapter 创建触达服务适配器
func NewReachAdapter() *ReachAdapter {
	return &ReachAdapter{}
}

// Send 发送触达消息
func (a *ReachAdapter) Send(ctx context.Context, template string, userID int64, params map[string]interface{}) error {
	// 模拟发送触达消息
	fmt.Printf("[Reach] Sending message to user %d with template: %s, params: %v\n", userID, template, params)

	// TODO: 实际的触达逻辑
	// 1. 调用触达服务API
	// 2. 选择触达渠道（Push、短信、站内信等）
	// 3. 处理失败重试

	return nil
}

// NotificationAdapter 通知服务适配器
type NotificationAdapter struct {
	reachAdapter *ReachAdapter
}

// NewNotificationAdapter 创建通知服务适配器
func NewNotificationAdapter(reachAdapter *ReachAdapter) *NotificationAdapter {
	return &NotificationAdapter{
		reachAdapter: reachAdapter,
	}
}

// SendTaskCompletedNotification 发送任务完成通知
func (a *NotificationAdapter) SendTaskCompletedNotification(ctx context.Context, userID int64, taskDetail *entity.ActUserTaskDetail) error {
	params := map[string]interface{}{
		"task_id":      taskDetail.TaskID,
		"reward_value": taskDetail.RewardValue,
	}

	return a.reachAdapter.Send(ctx, "task_completed", userID, params)
}

// SendTaskProgressNotification 发送任务进度通知
func (a *NotificationAdapter) SendTaskProgressNotification(ctx context.Context, userID int64, task *entity.ActUserTask) error {
	params := map[string]interface{}{
		"task_id":  task.ID,
		"progress": task.Progress,
		"target":   task.Target,
	}

	return a.reachAdapter.Send(ctx, "task_progress", userID, params)
}

