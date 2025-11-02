package output

import (
	"context"
	"mini-sirus/internal/domain/entity"
)

// NotificationService 通知服务输出端口
// 定义通知服务的抽象接口，具体实现在 adapter 层
type NotificationService interface {
	// SendTaskCompletedNotification 发送任务完成通知
	SendTaskCompletedNotification(ctx context.Context, userID int64, taskDetail *entity.ActUserTaskDetail) error

	// SendTaskProgressNotification 发送任务进度通知
	SendTaskProgressNotification(ctx context.Context, userID int64, task *entity.ActUserTask) error
}

// ReachService 触达服务输出端口
type ReachService interface {
	// Send 发送触达消息
	// template: 消息模板
	// userID: 用户ID
	// params: 额外参数
	Send(ctx context.Context, template string, userID int64, params map[string]interface{}) error
}

