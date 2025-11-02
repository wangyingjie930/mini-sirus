package input

import (
	"context"
	"mini-sirus/internal/usecase/dto"
)

// TaskService 任务服务输入端口
// 定义对外提供的任务服务接口
type TaskService interface {
	// TriggerTask 触发任务
	// 根据业务事件触发相应的任务检查和完成逻辑
	TriggerTask(ctx context.Context, input dto.TriggerTaskInput) error

	// CreateTask 创建任务
	CreateTask(ctx context.Context, input dto.CreateTaskInput) (*dto.TaskOutput, error)

	// QueryTask 查询任务
	QueryTask(ctx context.Context, input dto.QueryTaskInput) (*dto.TaskOutput, error)

	// QueryTasksByUser 查询用户的任务列表
	QueryTasksByUser(ctx context.Context, userID int64) ([]*dto.TaskOutput, error)
}

