package memory

import (
	"context"
	"errors"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/valueobject"
	"sync"
	"time"
)

// TaskRepositoryMemory 任务仓储内存实现
type TaskRepositoryMemory struct {
	mu      sync.RWMutex
	tasks   map[int64]*entity.ActUserTask
	idGen   int64
}

// NewTaskRepositoryMemory 创建内存任务仓储
func NewTaskRepositoryMemory() *TaskRepositoryMemory {
	return &TaskRepositoryMemory{
		tasks: make(map[int64]*entity.ActUserTask),
		idGen: 1000,
	}
}

// Create 创建任务
func (r *TaskRepositoryMemory) Create(ctx context.Context, task *entity.ActUserTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.idGen++
	task.ID = r.idGen
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// 复制一份存储，避免外部修改
	taskCopy := *task
	r.tasks[task.ID] = &taskCopy

	return nil
}

// Update 更新任务
func (r *TaskRepositoryMemory) Update(ctx context.Context, task *entity.ActUserTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return errors.New("task not found")
	}

	task.UpdatedAt = time.Now()
	taskCopy := *task
	r.tasks[task.ID] = &taskCopy

	return nil
}

// GetByID 根据ID获取任务
func (r *TaskRepositoryMemory) GetByID(ctx context.Context, taskID int64) (*entity.ActUserTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	taskCopy := *task
	return &taskCopy, nil
}

// ListByUserID 获取用户的任务列表
func (r *TaskRepositoryMemory) ListByUserID(ctx context.Context, userID int64) ([]*entity.ActUserTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.ActUserTask
	for _, task := range r.tasks {
		if task.UserID == userID {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}

	return result, nil
}

// ListByUserIDAndType 根据用户ID和任务类型获取任务列表
func (r *TaskRepositoryMemory) ListByUserIDAndType(ctx context.Context, userID int64, taskType valueobject.TaskType) ([]*entity.ActUserTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.ActUserTask
	for _, task := range r.tasks {
		if task.UserID == userID && task.TaskType == taskType {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}

	return result, nil
}

// UpdateProgress 更新任务进度
func (r *TaskRepositoryMemory) UpdateProgress(ctx context.Context, taskID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	task.UpdateProgress()
	return nil
}

