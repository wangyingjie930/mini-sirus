package task

import (
	"context"
	"errors"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/repository"
	"mini-sirus/internal/usecase/dto"
	"time"
)

// CreateTaskUseCase 创建任务用例
type CreateTaskUseCase struct {
	taskRepo repository.TaskRepository
}

// NewCreateTaskUseCase 创建任务用例构造函数
func NewCreateTaskUseCase(taskRepo repository.TaskRepository) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		taskRepo: taskRepo,
	}
}

// Execute 执行创建任务用例
func (uc *CreateTaskUseCase) Execute(ctx context.Context, input dto.CreateTaskInput) (*dto.TaskOutput, error) {
	// 验证输入
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// 创建任务实体
	task := &entity.ActUserTask{
		ActivityID:   input.ActivityID,
		TaskID:       input.TaskID,
		UserID:       input.UserID,
		TaskType:     input.TaskType,
		Status:       entity.TaskStatusPending,
		Progress:     0,
		Target:       input.Target,
		TaskCondExpr: input.TaskCondExpr,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 验证实体
	if !task.IsValid() {
		return nil, errors.New("invalid task entity")
	}

	// 保存任务
	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 转换为输出DTO
	return uc.toTaskOutput(task), nil
}

// validateInput 验证输入
func (uc *CreateTaskUseCase) validateInput(input dto.CreateTaskInput) error {
	if input.ActivityID <= 0 {
		return errors.New("activity_id is required")
	}
	if input.TaskID <= 0 {
		return errors.New("task_id is required")
	}
	if input.UserID <= 0 {
		return errors.New("user_id is required")
	}
	if input.Target <= 0 {
		return errors.New("target must be greater than 0")
	}
	if !input.TaskType.IsValid() {
		return errors.New("invalid task type")
	}
	if input.TaskCondExpr == "" {
		return errors.New("task_cond_expr is required")
	}
	return nil
}

// toTaskOutput 转换为输出DTO
func (uc *CreateTaskUseCase) toTaskOutput(task *entity.ActUserTask) *dto.TaskOutput {
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

