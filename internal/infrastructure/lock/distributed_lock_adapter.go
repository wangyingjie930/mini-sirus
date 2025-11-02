package infrastructure

import (
	"context"
	"mini-sirus/internal/usecase/port/output"
)

// DistributedLockAdapter 分布式锁适配器
type DistributedLockAdapter struct {
	memoryLock *MemoryLock
}

// NewDistributedLockAdapter 创建分布式锁适配器
func NewDistributedLockAdapter(memoryLock *MemoryLock) *DistributedLockAdapter {
	return &DistributedLockAdapter{
		memoryLock: memoryLock,
	}
}

// 确保实现了接口
var _ output.DistributedLock = (*DistributedLockAdapter)(nil)

// Lock 加锁
func (a *DistributedLockAdapter) Lock(ctx context.Context, key string, ttl int) (string, error) {
	return a.memoryLock.Lock(key, ttl)
}

// Unlock 解锁
func (a *DistributedLockAdapter) Unlock(ctx context.Context, key string, lockID string) error {
	return a.memoryLock.Unlock(key, lockID)
}

// TryLock 尝试加锁（非阻塞）
func (a *DistributedLockAdapter) TryLock(ctx context.Context, key string, ttl int) (bool, string, error) {
	return a.memoryLock.TryLock(key, ttl)
}

