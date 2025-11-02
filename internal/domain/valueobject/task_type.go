package valueobject

// TaskType 任务类型值对象
type TaskType string

const (
	TaskTypePublishTimes TaskType = "publish_times" // 发布篇数任务
	TaskTypeShareTimes   TaskType = "share_times"   // 分享次数任务
	TaskTypeLikeTimes    TaskType = "like_times"    // 点赞次数任务
	TaskTypeCommentTimes TaskType = "comment_times" // 评论次数任务
	TaskTypeCheckin      TaskType = "checkin"       // 签到任务
)

// IsValid 验证任务类型是否有效
func (t TaskType) IsValid() bool {
	switch t {
	case TaskTypePublishTimes, TaskTypeShareTimes, TaskTypeLikeTimes,
		TaskTypeCommentTimes, TaskTypeCheckin:
		return true
	default:
		return false
	}
}

// String 返回任务类型的字符串表示
func (t TaskType) String() string {
	return string(t)
}

// Equals 比较两个任务类型是否相等
func (t TaskType) Equals(other TaskType) bool {
	return t == other
}

