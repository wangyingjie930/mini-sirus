package dto

import (
	"mini-sirus/internal/domain/valueobject"
)

// TaskModeDTO 任务模式数据传输对象接口
type TaskModeDTO interface {
	// GetTaskType 获取任务类型
	GetTaskType() valueobject.TaskType

	// GetUserID 获取用户ID
	GetUserID() int64

	// GetUniqueFlag 获取唯一标识
	GetUniqueFlag() string

	// GetExpressionArguments 获取表达式参数
	GetExpressionArguments() valueobject.ExpressionArguments

	// GetExpressionFunctions 获取表达式函数（函数名列表）
	GetExpressionFunctions() []string
}

// PublishEventDTO 发布事件DTO
type PublishEventDTO struct {
	UserID       int64
	ContentID    int64
	TopicIDs     []uint64
	LikeCount    int
	CommentCount int
	IsAudited    bool
	AuditStatus  int
}

// GetTaskType 实现 TaskModeDTO 接口
func (p *PublishEventDTO) GetTaskType() valueobject.TaskType {
	return valueobject.TaskTypePublishTimes
}

// GetUserID 实现 TaskModeDTO 接口
func (p *PublishEventDTO) GetUserID() int64 {
	return p.UserID
}

// GetUniqueFlag 实现 TaskModeDTO 接口
func (p *PublishEventDTO) GetUniqueFlag() string {
	// 使用内容ID作为唯一标识
	return ""
}

// GetExpressionArguments 实现 TaskModeDTO 接口
func (p *PublishEventDTO) GetExpressionArguments() valueobject.ExpressionArguments {
	return valueobject.ExpressionArguments{
		"user_id":          float64(p.UserID),
		"content_id":       float64(p.ContentID),
		"tag_ids":          p.TopicIDs,
		"required_tag_ids": []uint64{1001, 1002}, // 任务要求的话题ID列表
		"like_count":       float64(p.LikeCount),
		"comment_count":    float64(p.CommentCount),
		"is_audited":       p.IsAudited,
		"audit_status":     float64(p.AuditStatus),
	}
}

// GetExpressionFunctions 实现 TaskModeDTO 接口
func (p *PublishEventDTO) GetExpressionFunctions() []string {
	return []string{"WITH_ANY_TOPIC", "LIKE_COUNT_GTE", "IS_AUDITED"}
}

// CheckinEventDTO 签到事件DTO
type CheckinEventDTO struct {
	UserID int64
	Date   string
}

// GetTaskType 实现 TaskModeDTO 接口
func (c *CheckinEventDTO) GetTaskType() valueobject.TaskType {
	return valueobject.TaskTypeCheckin
}

// GetUserID 实现 TaskModeDTO 接口
func (c *CheckinEventDTO) GetUserID() int64 {
	return c.UserID
}

// GetUniqueFlag 实现 TaskModeDTO 接口
func (c *CheckinEventDTO) GetUniqueFlag() string {
	return ""
}

// GetExpressionArguments 实现 TaskModeDTO 接口
func (c *CheckinEventDTO) GetExpressionArguments() valueobject.ExpressionArguments {
	return valueobject.ExpressionArguments{
		"user_id": float64(c.UserID),
		"date":    c.Date,
	}
}

// GetExpressionFunctions 实现 TaskModeDTO 接口
func (c *CheckinEventDTO) GetExpressionFunctions() []string {
	return []string{"IS_TODAY"}
}

