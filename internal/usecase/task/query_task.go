package task

import (
	"context"
	"errors"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/repository"
	"mini-sirus/internal/usecase/dto"
	"time"
)

// QueryTaskUseCase 查询任务用例
type QueryTaskUseCase struct {
	taskRepo repository.TaskRepository
}

// NewQueryTaskUseCase 创建查询任务用例
func NewQueryTaskUseCase(taskRepo repository.TaskRepository) *QueryTaskUseCase {
	return &QueryTaskUseCase{
		taskRepo: taskRepo,
	}
}

// Execute 执行查询任务用例（单个任务）
func (uc *QueryTaskUseCase) Execute(ctx context.Context, input dto.QueryTaskInput) (*dto.TaskOutput, error) {
	if input.TaskID <= 0 {
		return nil, errors.New("task_id is required")
	}

	task, err := uc.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, err
	}

	return uc.toTaskOutput(task), nil
}

// ExecuteList 执行查询任务用例（用户任务列表）
func (uc *QueryTaskUseCase) ExecuteList(ctx context.Context, userID int64) ([]*dto.TaskOutput, error) {
	if userID <= 0 {
		return nil, errors.New("user_id is required")
	}

	tasks, err := uc.taskRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.TaskOutput, 0, len(tasks))
	for _, task := range tasks {
		outputs = append(outputs, uc.toTaskOutput(task))
	}

	return outputs, nil
}

// toTaskOutput 转换为输出DTO
func (uc *QueryTaskUseCase) toTaskOutput(task *entity.ActUserTask) *dto.TaskOutput {
	return &dto.TaskOutput{
		ID:           task.ID,
		ActivityID:   task.ActivityID,
		TaskID:       task.TaskID,
		UserID:       task.UserID,
		TaskType:     task.TaskType,
		Status:       task.Status.String(),
		Progress:     task.Progress,
		Target:       task.Target,
		TaskCondExpr: task.TaskCondExpr,
		CreatedAt:    task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    task.UpdatedAt.Format(time.RFC3339),
	}
}

