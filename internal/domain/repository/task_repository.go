package repository

import (
	"context"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/valueobject"
)

// TaskRepository 任务仓储接口
// 定义任务数据访问的抽象，具体实现在 adapter 层
type TaskRepository interface {
	// Create 创建任务
	Create(ctx context.Context, task *entity.ActUserTask) error

	// Update 更新任务
	Update(ctx context.Context, task *entity.ActUserTask) error

	// GetByID 根据ID获取任务
	GetByID(ctx context.Context, taskID int64) (*entity.ActUserTask, error)

	// ListByUserID 获取用户的任务列表
	ListByUserID(ctx context.Context, userID int64) ([]*entity.ActUserTask, error)

	// ListByUserIDAndType 根据用户ID和任务类型获取任务列表
	ListByUserIDAndType(ctx context.Context, userID int64, taskType valueobject.TaskType) ([]*entity.ActUserTask, error)

	// UpdateProgress 更新任务进度
	UpdateProgress(ctx context.Context, taskID int64) error
}

// TaskDetailRepository 任务明细仓储接口
type TaskDetailRepository interface {
	// Create 创建任务明细
	Create(ctx context.Context, detail *entity.ActUserTaskDetail) error

	// GetByID 根据ID获取任务明细
	GetByID(ctx context.Context, detailID int64) (*entity.ActUserTaskDetail, error)

	// ListByTaskID 根据任务ID获取明细列表
	ListByTaskID(ctx context.Context, taskID int64) ([]*entity.ActUserTaskDetail, error)

	// ExistsByUniqueFlag 判断唯一标识是否已存在
	ExistsByUniqueFlag(ctx context.Context, uniqueFlag string) (bool, error)
}

