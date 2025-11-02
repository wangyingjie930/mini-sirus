package output

import (
	"context"
	"mini-sirus/internal/domain/entity"
	"time"
)

// RiskCheckService 风控检查服务输出端口
type RiskCheckService interface {
	// CheckUserBehavior 检查用户行为异常
	CheckUserBehavior(ctx context.Context, userID int64, detail *entity.ActUserTaskDetail) error

	// CheckTaskFrequency 检查任务完成频率
	CheckTaskFrequency(ctx context.Context, userID, taskID int64) error

	// CheckDeviceFingerprint 检查设备指纹（简化版本，实际需要从请求上下文获取设备信息）
	CheckDeviceFingerprint(ctx context.Context, userID int64, detail *entity.ActUserTaskDetail) error

	// RecordTaskCompletion 记录任务完成事件（用于频率统计）
	RecordTaskCompletion(ctx context.Context, userID, taskID int64, timestamp time.Time) error

	// IsUserBlacklisted 检查用户是否在黑名单中
	IsUserBlacklisted(ctx context.Context, userID int64) (bool, error)

	// AddToBlacklist 将用户加入黑名单
	AddToBlacklist(ctx context.Context, userID int64, reason string) error
}

// UserBehaviorRecord 用户行为记录
type UserBehaviorRecord struct {
	UserID    int64
	Action    string
	TaskID    int64
	Timestamp time.Time
}

// TaskCompletionRecord 任务完成记录
type TaskCompletionRecord struct {
	UserID    int64
	TaskID    int64
	Timestamp time.Time
}

