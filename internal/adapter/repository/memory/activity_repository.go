package memory

import (
	"context"
	"errors"
	"mini-sirus/internal/domain/entity"
	"sync"
	"time"
)

// ActivityRepositoryMemory 活动仓储内存实现
type ActivityRepositoryMemory struct {
	mu         sync.RWMutex
	activities map[int64]*entity.ActActivity
	idGen      int64
}

// NewActivityRepositoryMemory 创建内存活动仓储
func NewActivityRepositoryMemory() *ActivityRepositoryMemory {
	return &ActivityRepositoryMemory{
		activities: make(map[int64]*entity.ActActivity),
		idGen:      3000,
	}
}

// Create 创建活动
func (r *ActivityRepositoryMemory) Create(ctx context.Context, activity *entity.ActActivity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idGen++
	activity.ID = r.idGen

	activityCopy := *activity
	r.activities[activity.ID] = &activityCopy

	return nil
}

// Update 更新活动
func (r *ActivityRepositoryMemory) Update(ctx context.Context, activity *entity.ActActivity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.activities[activity.ID]; !exists {
		return errors.New("activity not found")
	}

	activityCopy := *activity
	r.activities[activity.ID] = &activityCopy

	return nil
}

// GetByID 根据ID获取活动
func (r *ActivityRepositoryMemory) GetByID(ctx context.Context, activityID int64) (*entity.ActActivity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activity, exists := r.activities[activityID]
	if !exists {
		return nil, errors.New("activity not found")
	}

	activityCopy := *activity
	return &activityCopy, nil
}

// ListActive 获取活动中的活动列表
func (r *ActivityRepositoryMemory) ListActive(ctx context.Context) ([]*entity.ActActivity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.ActActivity
	now := time.Now()

	for _, activity := range r.activities {
		if activity.Status == entity.ActivityStatusActive &&
			now.After(activity.StartTime) &&
			now.Before(activity.EndTime) {
			activityCopy := *activity
			result = append(result, &activityCopy)
		}
	}

	return result, nil
}

