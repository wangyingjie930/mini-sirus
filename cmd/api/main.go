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

	// 注册观察者
	checkinObserver := observer.NewCheckinReachObserver(reachAdapter)
	riskCheckObserver := observer.NewRiskCheckObserver()
	observerRegistry.Register(checkinObserver)
	observerRegistry.Register(riskCheckObserver)

	// 初始化用例层
	triggerTaskUC := task.NewTriggerTaskUseCase(
		taskRepo,
		taskDetailRepo,
		ruleEngine,
		observerRegistry,
		distributedLock,
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

