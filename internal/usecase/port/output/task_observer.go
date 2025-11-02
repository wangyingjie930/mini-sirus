package output

import (
	"context"
	"mini-sirus/internal/domain/entity"
)

// TaskObserver 任务观察者输出端口
// 定义任务事件观察者的抽象接口
type TaskObserver interface {
	// OnTaskDetailCreated 当任务明细创建时
	OnTaskDetailCreated(ctx context.Context, detail *entity.ActUserTaskDetail) error

	// OnTaskCompleted 当任务完成时
	OnTaskCompleted(ctx context.Context, task *entity.ActUserTask) error

	// GetObserverName 获取观察者名称（用于标识）
	GetObserverName() string
}

// TaskObserverRegistry 任务观察者注册表
type TaskObserverRegistry interface {
	// Register 注册观察者
	Register(observer TaskObserver)

	// Unregister 注销观察者
	Unregister(observerName string)

	// Notify 通知所有观察者
	Notify(ctx context.Context, detail *entity.ActUserTaskDetail) error
}

