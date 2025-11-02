package memory

import (
	"context"
	"errors"
	"mini-sirus/internal/domain/entity"
	"sync"
	"time"
)

// TaskDetailRepositoryMemory 任务明细仓储内存实现
type TaskDetailRepositoryMemory struct {
	mu      sync.RWMutex
	details map[int64]*entity.ActUserTaskDetail
	idGen   int64
}

// NewTaskDetailRepositoryMemory 创建内存任务明细仓储
func NewTaskDetailRepositoryMemory() *TaskDetailRepositoryMemory {
	return &TaskDetailRepositoryMemory{
		details: make(map[int64]*entity.ActUserTaskDetail),
		idGen:   2000,
	}
}

// Create 创建任务明细
func (r *TaskDetailRepositoryMemory) Create(ctx context.Context, detail *entity.ActUserTaskDetail) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idGen++
	detail.ID = r.idGen
	detail.CreatedAt = time.Now()
	detail.UpdatedAt = time.Now()

	detailCopy := *detail
	r.details[detail.ID] = &detailCopy

	return nil
}

// GetByID 根据ID获取任务明细
func (r *TaskDetailRepositoryMemory) GetByID(ctx context.Context, detailID int64) (*entity.ActUserTaskDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	detail, exists := r.details[detailID]
	if !exists {
		return nil, errors.New("task detail not found")
	}

	detailCopy := *detail
	return &detailCopy, nil
}

// ListByTaskID 根据任务ID获取明细列表
func (r *TaskDetailRepositoryMemory) ListByTaskID(ctx context.Context, taskID int64) ([]*entity.ActUserTaskDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.ActUserTaskDetail
	for _, detail := range r.details {
		if detail.TaskID == taskID {
			detailCopy := *detail
			result = append(result, &detailCopy)
		}
	}

	return result, nil
}

// ExistsByUniqueFlag 判断唯一标识是否已存在
func (r *TaskDetailRepositoryMemory) ExistsByUniqueFlag(ctx context.Context, uniqueFlag string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, detail := range r.details {
		if detail.UniqueFlag == uniqueFlag {
			return true, nil
		}
	}

	return false, nil
}

