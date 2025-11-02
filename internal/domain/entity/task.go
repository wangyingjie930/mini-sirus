package entity

import (
	"mini-sirus/internal/domain/valueobject"
	"time"
)

// ActUserTask 用户任务实体
// 代表用户参与的活动任务，包含任务进度和状态
type ActUserTask struct {
	ID           int64
	ActivityID   int64
	TaskID       int64
	UserID       int64
	TaskType     valueobject.TaskType // 任务类型
	Status       TaskStatus
	Progress     int
	Target       int
	TaskCondExpr string // 任务条件表达式
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsCompleted 判断任务是否已完成
func (t *ActUserTask) IsCompleted() bool {
	return t.Status == TaskStatusDone
}

// IsPending 判断任务是否进行中
func (t *ActUserTask) IsPending() bool {
	return t.Status == TaskStatusPending
}

// CanProgress 判断任务是否可以更新进度
func (t *ActUserTask) CanProgress() bool {
	return t.IsPending() && t.Progress < t.Target
}

// UpdateProgress 更新任务进度
func (t *ActUserTask) UpdateProgress() {
	if !t.CanProgress() {
		return
	}

	t.Progress++
	if t.Progress >= t.Target {
		t.Status = TaskStatusDone
	}
	t.UpdatedAt = time.Now()
}

// IsValid 验证任务实体是否有效
func (t *ActUserTask) IsValid() bool {
	return t.ActivityID > 0 &&
		t.TaskID > 0 &&
		t.UserID > 0 &&
		t.Target > 0 &&
		t.TaskCondExpr != ""
}

// IsExpired 判断任务是否过期（简化版本，实际应关联活动）
func (t *ActUserTask) IsExpired(validDays int) bool {
	return time.Since(t.CreatedAt) > time.Duration(validDays)*24*time.Hour
}

// ActUserTaskDetail 用户任务明细实体
// 代表任务的每次完成记录
type ActUserTaskDetail struct {
	ID          int64
	TaskID      int64
	UserID      int64
	Status      TaskDetailStatus
	UniqueFlag  string // 唯一标识，防止重复
	RewardValue int    // 激励值
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsCompleted 判断明细是否已完成
func (d *ActUserTaskDetail) IsCompleted() bool {
	return d.Status == TaskDetailStatusDone
}

// IsPending 判断明细是否进行中
func (d *ActUserTaskDetail) IsPending() bool {
	return d.Status == TaskDetailStatusPending
}

// Complete 完成任务明细
func (d *ActUserTaskDetail) Complete() {
	d.Status = TaskDetailStatusDone
	d.UpdatedAt = time.Now()
}

// ActActivity 活动实体
type ActActivity struct {
	ID        int64
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Status    ActivityStatus
}

// IsActive 判断活动是否激活
func (a *ActActivity) IsActive() bool {
	now := time.Now()
	return a.Status == ActivityStatusActive &&
		now.After(a.StartTime) &&
		now.Before(a.EndTime)
}

// IsInTimeRange 判断是否在活动时间范围内
func (a *ActActivity) IsInTimeRange() bool {
	now := time.Now()
	return now.After(a.StartTime) && now.Before(a.EndTime)
}

