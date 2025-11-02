package memory

import (
	"context"
	"errors"
	"fmt"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/usecase/port/output"
	"sync"
	"time"
)

// RiskCheckServiceMemory 风控检查服务内存实现
type RiskCheckServiceMemory struct {
	mu sync.RWMutex

	// 用户行为记录
	userBehaviors map[int64][]output.UserBehaviorRecord

	// 任务完成记录
	taskCompletions map[int64][]output.TaskCompletionRecord

	// 黑名单
	blacklist map[int64]string // userID -> reason

	// 设备指纹记录 (userID -> deviceIDs)
	userDevices map[int64]map[string]bool

	// 设备关联用户 (deviceID -> userIDs)
	deviceUsers map[string]map[int64]bool
}

// NewRiskCheckServiceMemory 创建内存风控服务
func NewRiskCheckServiceMemory() *RiskCheckServiceMemory {
	return &RiskCheckServiceMemory{
		userBehaviors:   make(map[int64][]output.UserBehaviorRecord),
		taskCompletions: make(map[int64][]output.TaskCompletionRecord),
		blacklist:       make(map[int64]string),
		userDevices:     make(map[int64]map[string]bool),
		deviceUsers:     make(map[string]map[int64]bool),
	}
}

// CheckUserBehavior 检查用户行为异常
func (r *RiskCheckServiceMemory) CheckUserBehavior(ctx context.Context, userID int64, detail *entity.ActUserTaskDetail) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	behaviors := r.userBehaviors[userID]
	if len(behaviors) == 0 {
		return nil
	}

	// 1. 检查短时间内是否有大量操作（最近1分钟内）
	recentCount := 0
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	for _, behavior := range behaviors {
		if behavior.Timestamp.After(oneMinuteAgo) {
			recentCount++
		}
	}

	// 如果1分钟内操作超过10次，判定为异常
	if recentCount > 10 {
		return fmt.Errorf("用户行为异常: 1分钟内操作次数过多(%d次)", recentCount)
	}

	// 2. 检查操作时间间隔是否过于规律（机器人特征）
	if len(behaviors) >= 5 {
		recentBehaviors := behaviors[len(behaviors)-5:]
		intervals := make([]float64, 0)

		for i := 1; i < len(recentBehaviors); i++ {
			interval := recentBehaviors[i].Timestamp.Sub(recentBehaviors[i-1].Timestamp).Seconds()
			intervals = append(intervals, interval)
		}

		// 计算时间间隔的方差，如果方差很小（< 0.1秒），说明过于规律
		if len(intervals) > 0 {
			variance := calculateVariance(intervals)
			if variance < 0.1 {
				return fmt.Errorf("用户行为异常: 操作时间间隔过于规律(方差: %.4f)", variance)
			}
		}
	}

	return nil
}

// CheckTaskFrequency 检查任务完成频率
func (r *RiskCheckServiceMemory) CheckTaskFrequency(ctx context.Context, userID, taskID int64) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	completions := r.taskCompletions[userID]
	if len(completions) == 0 {
		return nil
	}

	now := time.Now()

	// 1. 检查1小时内同一任务的完成次数
	oneHourAgo := now.Add(-1 * time.Hour)
	taskCount := 0
	for _, completion := range completions {
		if completion.Timestamp.After(oneHourAgo) && completion.TaskID == taskID {
			taskCount++
		}
	}

	if taskCount >= 10 {
		return fmt.Errorf("任务完成频率过高: 1小时内完成同一任务%d次", taskCount)
	}

	// 2. 检查24小时内所有任务的完成次数
	oneDayAgo := now.Add(-24 * time.Hour)
	totalCount := 0
	for _, completion := range completions {
		if completion.Timestamp.After(oneDayAgo) {
			totalCount++
		}
	}

	if totalCount >= 100 {
		return fmt.Errorf("任务完成频率过高: 24小时内完成%d次任务", totalCount)
	}

	// 3. 检查新用户是否异常活跃（注册后24小时内完成超过20个任务）
	// 这里简化处理，假设完成任务少于50次的都是新用户
	if len(completions) < 50 {
		if totalCount > 20 {
			return fmt.Errorf("新用户异常活跃: 24小时内完成%d次任务", totalCount)
		}
	}

	return nil
}

// CheckDeviceFingerprint 检查设备指纹
func (r *RiskCheckServiceMemory) CheckDeviceFingerprint(ctx context.Context, userID int64, detail *entity.ActUserTaskDetail) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 这里使用 UniqueFlag 作为设备指纹的简化实现
	// 实际项目中应该从请求上下文中获取真实的设备指纹信息
	deviceID := detail.UniqueFlag
	if deviceID == "" {
		return nil
	}

	// 1. 检查单设备关联的账号数量
	if users, exists := r.deviceUsers[deviceID]; exists {
		accountCount := len(users)
		if accountCount > 5 {
			return fmt.Errorf("设备指纹异常: 单设备关联账号过多(%d个账号)", accountCount)
		}
	}

	// 2. 检查单用户使用的设备数量（频繁换设备也是异常行为）
	if devices, exists := r.userDevices[userID]; exists {
		deviceCount := len(devices)
		if deviceCount > 10 {
			return fmt.Errorf("设备指纹异常: 用户使用设备过多(%d个设备)", deviceCount)
		}
	}

	return nil
}

// RecordTaskCompletion 记录任务完成事件
func (r *RiskCheckServiceMemory) RecordTaskCompletion(ctx context.Context, userID, taskID int64, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 记录任务完成
	completion := output.TaskCompletionRecord{
		UserID:    userID,
		TaskID:    taskID,
		Timestamp: timestamp,
	}
	r.taskCompletions[userID] = append(r.taskCompletions[userID], completion)

	// 记录用户行为
	behavior := output.UserBehaviorRecord{
		UserID:    userID,
		Action:    "task_complete",
		TaskID:    taskID,
		Timestamp: timestamp,
	}
	r.userBehaviors[userID] = append(r.userBehaviors[userID], behavior)

	// 只保留最近1000条记录，避免内存无限增长
	if len(r.taskCompletions[userID]) > 1000 {
		r.taskCompletions[userID] = r.taskCompletions[userID][len(r.taskCompletions[userID])-1000:]
	}
	if len(r.userBehaviors[userID]) > 1000 {
		r.userBehaviors[userID] = r.userBehaviors[userID][len(r.userBehaviors[userID])-1000:]
	}

	return nil
}

// IsUserBlacklisted 检查用户是否在黑名单中
func (r *RiskCheckServiceMemory) IsUserBlacklisted(ctx context.Context, userID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.blacklist[userID]
	return exists, nil
}

// AddToBlacklist 将用户加入黑名单
func (r *RiskCheckServiceMemory) AddToBlacklist(ctx context.Context, userID int64, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if reason == "" {
		return errors.New("加入黑名单必须提供原因")
	}

	r.blacklist[userID] = reason
	return nil
}

// UpdateDeviceMapping 更新设备映射关系（内部辅助方法）
func (r *RiskCheckServiceMemory) UpdateDeviceMapping(userID int64, deviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新用户设备映射
	if r.userDevices[userID] == nil {
		r.userDevices[userID] = make(map[string]bool)
	}
	r.userDevices[userID][deviceID] = true

	// 更新设备用户映射
	if r.deviceUsers[deviceID] == nil {
		r.deviceUsers[deviceID] = make(map[int64]bool)
	}
	r.deviceUsers[deviceID][userID] = true
}

// calculateVariance 计算方差
func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// 计算平均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 计算方差
	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}

	return varianceSum / float64(len(values))
}

