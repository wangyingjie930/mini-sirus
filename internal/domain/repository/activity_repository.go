package repository

import (
	"context"
	"mini-sirus/internal/domain/entity"
)

// ActivityRepository 活动仓储接口
type ActivityRepository interface {
	// Create 创建活动
	Create(ctx context.Context, activity *entity.ActActivity) error

	// Update 更新活动
	Update(ctx context.Context, activity *entity.ActActivity) error

	// GetByID 根据ID获取活动
	GetByID(ctx context.Context, activityID int64) (*entity.ActActivity, error)

	// ListActive 获取活动中的活动列表
	ListActive(ctx context.Context) ([]*entity.ActActivity, error)
}

