package output

import "context"

// DistributedLock 分布式锁输出端口
// 定义分布式锁的抽象接口，具体实现在 infrastructure 层
type DistributedLock interface {
	// Lock 加锁
	// key: 锁的键
	// ttl: 锁的过期时间（秒）
	// 返回: 锁的标识和错误
	Lock(ctx context.Context, key string, ttl int) (string, error)

	// Unlock 解锁
	// key: 锁的键
	// lockID: 锁的标识
	Unlock(ctx context.Context, key string, lockID string) error

	// TryLock 尝试加锁（非阻塞）
	TryLock(ctx context.Context, key string, ttl int) (bool, string, error)
}

