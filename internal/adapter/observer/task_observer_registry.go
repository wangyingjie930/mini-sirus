package observer

import (
	"context"
	"fmt"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/usecase/port/output"
	"sync"
)

// TaskObserverRegistry 任务观察者注册表实现
type TaskObserverRegistry struct {
	mu        sync.RWMutex
	observers map[string]output.TaskObserver
}

// NewTaskObserverRegistry 创建任务观察者注册表
func NewTaskObserverRegistry() *TaskObserverRegistry {
	return &TaskObserverRegistry{
		observers: make(map[string]output.TaskObserver),
	}
}

// Register 注册观察者
func (r *TaskObserverRegistry) Register(observer output.TaskObserver) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := observer.GetObserverName()
	if name == "" {
		return
	}

	if _, exists := r.observers[name]; exists {
		return // 避免重复注册
	}

	r.observers[name] = observer
}

// Unregister 注销观察者
func (r *TaskObserverRegistry) Unregister(observerName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.observers, observerName)
}

// Notify 通知所有观察者
func (r *TaskObserverRegistry) Notify(ctx context.Context, detail *entity.ActUserTaskDetail) error {
	r.mu.RLock()
	observersCopy := make([]output.TaskObserver, 0, len(r.observers))
	for _, obs := range r.observers {
		observersCopy = append(observersCopy, obs)
	}
	r.mu.RUnlock()

	// 串行通知所有观察者
	for _, observer := range observersCopy {
		if err := observer.OnTaskDetailCreated(ctx, detail); err != nil {
			// 记录错误但继续执行
			fmt.Printf("[Observer] %s failed: %v\n", observer.GetObserverName(), err)
			return fmt.Errorf("observer %s failed: %w", observer.GetObserverName(), err)
		}
	}

	return nil
}

