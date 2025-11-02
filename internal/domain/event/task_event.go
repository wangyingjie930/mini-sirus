package event

import (
	"time"
)

// TaskCompleted 任务完成事件
type TaskCompleted struct {
	TaskID      int64
	UserID      int64
	ActivityID  int64
	CompletedAt time.Time
}

// TaskProgressUpdated 任务进度更新事件
type TaskProgressUpdated struct {
	TaskID      int64
	UserID      int64
	Progress    int
	Target      int
	UpdatedAt   time.Time
}

// TaskDetailCreated 任务明细创建事件
type TaskDetailCreated struct {
	DetailID    int64
	TaskID      int64
	UserID      int64
	UniqueFlag  string
	RewardValue int
	CreatedAt   time.Time
}

// PublishEvent 发布事件（业务事件）
type PublishEvent struct {
	UserID       int64
	ContentID    int64
	TopicIDs     []uint64
	LikeCount    int
	CommentCount int
	IsAudited    bool
	AuditStatus  int
	PublishedAt  time.Time
}

// CheckinEvent 签到事件（业务事件）
type CheckinEvent struct {
	UserID      int64
	CheckinDate string
	CheckinAt   time.Time
}

