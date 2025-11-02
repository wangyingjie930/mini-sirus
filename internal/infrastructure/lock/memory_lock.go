package infrastructure

import (
	"fmt"
	"sync"
)

// MemoryLock 内存分布式锁（模拟实现，仅用于单机测试）
type MemoryLock struct {
	mu    sync.Mutex
	locks map[string]string // key -> lockID
}

// NewMemoryLock 创建内存锁
func NewMemoryLock() *MemoryLock {
	return &MemoryLock{
		locks: make(map[string]string),
	}
}

// Lock 加锁
func (l *MemoryLock) Lock(key string, ttl int) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.locks[key]; exists {
		return "", fmt.Errorf("lock already exists for key: %s", key)
	}

	lockID := fmt.Sprintf("lock_%s_%d", key, len(l.locks)+1)
	l.locks[key] = lockID

	return lockID, nil
}

// Unlock 解锁
func (l *MemoryLock) Unlock(key string, lockID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	existingLockID, exists := l.locks[key]
	if !exists {
		return fmt.Errorf("lock not found for key: %s", key)
	}

	if existingLockID != lockID {
		return fmt.Errorf("lock id mismatch for key: %s", key)
	}

	delete(l.locks, key)
	return nil
}

// TryLock 尝试加锁（非阻塞）
func (l *MemoryLock) TryLock(key string, ttl int) (bool, string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.locks[key]; exists {
		return false, "", nil
	}

	lockID := fmt.Sprintf("lock_%s_%d", key, len(l.locks)+1)
	l.locks[key] = lockID

	return true, lockID, nil
}

