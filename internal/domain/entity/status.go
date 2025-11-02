package entity

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending TaskStatus = 0 // 进行中
	TaskStatusDone    TaskStatus = 1 // 已完成
)

// String 返回状态的字符串表示
func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "pending"
	case TaskStatusDone:
		return "done"
	default:
		return "unknown"
	}
}

// TaskDetailStatus 任务明细状态
type TaskDetailStatus int

const (
	TaskDetailStatusPending TaskDetailStatus = 0 // 进行中
	TaskDetailStatusDone    TaskDetailStatus = 1 // 已完成
)

// String 返回状态的字符串表示
func (s TaskDetailStatus) String() string {
	switch s {
	case TaskDetailStatusPending:
		return "pending"
	case TaskDetailStatusDone:
		return "done"
	default:
		return "unknown"
	}
}

// ActivityStatus 活动状态
type ActivityStatus int

const (
	ActivityStatusInactive ActivityStatus = 0 // 未激活
	ActivityStatusActive   ActivityStatus = 1 // 激活中
	ActivityStatusExpired  ActivityStatus = 2 // 已过期
)

// String 返回状态的字符串表示
func (s ActivityStatus) String() string {
	switch s {
	case ActivityStatusInactive:
		return "inactive"
	case ActivityStatusActive:
		return "active"
	case ActivityStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

