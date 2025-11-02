package dto

import (
	"mini-sirus/internal/domain/valueobject"
)

// TriggerTaskInput 触发任务输入
type TriggerTaskInput struct {
	TaskMode TaskModeDTO
}

// CreateTaskInput 创建任务输入
type CreateTaskInput struct {
	ActivityID   int64
	TaskID       int64
	UserID       int64
	Target       int
	TaskType     valueobject.TaskType
	TaskCondExpr string
}

// QueryTaskInput 查询任务输入
type QueryTaskInput struct {
	UserID   int64
	TaskID   int64
	TaskType valueobject.TaskType
}

// TaskOutput 任务输出
type TaskOutput struct {
	ID           int64                 `json:"id"`
	ActivityID   int64                 `json:"activity_id"`
	TaskID       int64                 `json:"task_id"`
	UserID       int64                 `json:"user_id"`
	TaskType     valueobject.TaskType  `json:"task_type"`
	Status       string                `json:"status"`
	Progress     int                   `json:"progress"`
	Target       int                   `json:"target"`
	TaskCondExpr string                `json:"task_cond_expr"`
	CreatedAt    string                `json:"created_at"`
	UpdatedAt    string                `json:"updated_at"`
}

// TaskDetailOutput 任务明细输出
type TaskDetailOutput struct {
	ID          int64  `json:"id"`
	TaskID      int64  `json:"task_id"`
	UserID      int64  `json:"user_id"`
	Status      string `json:"status"`
	UniqueFlag  string `json:"unique_flag"`
	RewardValue int    `json:"reward_value"`
	CreatedAt   string `json:"created_at"`
}

