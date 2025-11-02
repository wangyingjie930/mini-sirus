package main

import (
	"fmt"
	"mini-sirus/internal/adapter/notification"
	"mini-sirus/internal/adapter/observer"
	"mini-sirus/internal/adapter/repository/memory"
	"mini-sirus/internal/adapter/rule_engine"
	"mini-sirus/internal/infrastructure/config"
	infrastructure "mini-sirus/internal/infrastructure/lock"
	"mini-sirus/internal/infrastructure/logger"
	"mini-sirus/internal/interface/http/handler"
	"mini-sirus/internal/interface/http/router"
	"mini-sirus/internal/usecase/task"
	"net/http"
)

func main() {
	// 加载配置
	cfg := config.NewDefaultConfig()
	log := logger.NewSimpleLogger("mini-sirus-api")

	log.Info("Starting Mini-Sirus API Server...")

	// 初始化仓储层
	taskRepo := memory.NewTaskRepositoryMemory()
	taskDetailRepo := memory.NewTaskDetailRepositoryMemory()

	// 初始化适配器层
	ruleEngine := rule_engine.NewGovaluateAdapter()
	observerRegistry := observer.NewTaskObserverRegistry()
	memLock := infrastructure.NewMemoryLock()
	distributedLock := infrastructure.NewDistributedLockAdapter(memLock)
	reachAdapter := notification.NewReachAdapter()
	riskCheckService := memory.NewRiskCheckServiceMemory()

	// 注册观察者（仅注册适合异步执行的观察者）
	// 风控服务不应该作为观察者，而应该在用例层同步执行
	checkinObserver := observer.NewCheckinReachObserver(reachAdapter)
	observerRegistry.Register(checkinObserver)

	// 初始化用例层
	// 风控服务作为依赖注入到 TriggerTaskUseCase
	triggerTaskUC := task.NewTriggerTaskUseCase(
		taskRepo,
		taskDetailRepo,
		ruleEngine,
		observerRegistry,
		distributedLock,
		riskCheckService, // 风控服务作为依赖注入，在任务完成前同步执行
	)
	createTaskUC := task.NewCreateTaskUseCase(taskRepo)
	queryTaskUC := task.NewQueryTaskUseCase(taskRepo)

	// 初始化接口层
	taskHandler := handler.NewTaskHandler(triggerTaskUC, createTaskUC, queryTaskUC)
	r := router.NewRouter(taskHandler)

	// 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info(fmt.Sprintf("Server listening on %s", addr))

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("Server failed to start", "error", err)
		panic(err)
	}
}

