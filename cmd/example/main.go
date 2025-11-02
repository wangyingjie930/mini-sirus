package main

import (
	"context"
	"fmt"
	"log"
	"mini-sirus/internal/adapter/notification"
	"mini-sirus/internal/adapter/observer"
	"mini-sirus/internal/adapter/repository/memory"
	"mini-sirus/internal/adapter/rule_engine"
	"mini-sirus/internal/domain/entity"
	"mini-sirus/internal/domain/valueobject"
	"mini-sirus/internal/infrastructure/config"
	infrastructure "mini-sirus/internal/infrastructure/lock"
	"mini-sirus/internal/infrastructure/logger"
	"mini-sirus/internal/usecase/dto"
	"mini-sirus/internal/usecase/task"
	"time"
)

// Container 依赖注入容器
type Container struct {
	// Repositories
	TaskRepo       *memory.TaskRepositoryMemory
	TaskDetailRepo *memory.TaskDetailRepositoryMemory
	ActivityRepo   *memory.ActivityRepositoryMemory

	// Adapters
	RuleEngine       *rule_engine.GovaluateAdapter
	ObserverRegistry *observer.TaskObserverRegistry
	DistributedLock  *infrastructure.DistributedLockAdapter
	ReachAdapter     *notification.ReachAdapter

	// Use Cases
	TriggerTaskUC *task.TriggerTaskUseCase
	CreateTaskUC  *task.CreateTaskUseCase
	QueryTaskUC   *task.QueryTaskUseCase

	// Infrastructure
	Config *config.Config
	Logger logger.Logger
}

// NewContainer 创建依赖注入容器
func NewContainer() *Container {
	// 配置和日志
	cfg := config.NewDefaultConfig()
	log := logger.NewSimpleLogger("mini-sirus")

	// 仓储层
	taskRepo := memory.NewTaskRepositoryMemory()
	taskDetailRepo := memory.NewTaskDetailRepositoryMemory()
	activityRepo := memory.NewActivityRepositoryMemory()

	// 适配器层
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

	// 用例层
	triggerTaskUC := task.NewTriggerTaskUseCase(
		taskRepo,
		taskDetailRepo,
		ruleEngine,
		observerRegistry,
		distributedLock,
	)
	createTaskUC := task.NewCreateTaskUseCase(taskRepo)
	queryTaskUC := task.NewQueryTaskUseCase(taskRepo)

	return &Container{
		TaskRepo:         taskRepo,
		TaskDetailRepo:   taskDetailRepo,
		ActivityRepo:     activityRepo,
		RuleEngine:       ruleEngine,
		ObserverRegistry: observerRegistry,
		DistributedLock:  distributedLock,
		ReachAdapter:     reachAdapter,
		TriggerTaskUC:    triggerTaskUC,
		CreateTaskUC:     createTaskUC,
		QueryTaskUC:      queryTaskUC,
		Config:           cfg,
		Logger:           log,
	}
}

func main() {
	fmt.Println("=== Mini-Sirus Clean Architecture Example ===\n")

	// 初始化容器
	container := NewContainer()
	container.Logger.Info("Application started")

	ctx := context.Background()

	// 示例1: 创建任务
	fmt.Println("--- Example 1: Create Task ---")
	createInput := dto.CreateTaskInput{
		ActivityID:   1,
		TaskID:       100,
		UserID:       12345,
		Target:       3,
		TaskType:     valueobject.TaskTypePublishTimes,
		TaskCondExpr: "WITH_ANY_TOPIC(tag_ids, required_tag_ids) && LIKE_COUNT_GTE(like_count, 10) && IS_AUDITED(is_audited)",
	}

	taskOutput, err := container.CreateTaskUC.Execute(ctx, createInput)
	if err != nil {
		log.Fatalf("Create task failed: %v", err)
	}
	fmt.Printf("Task created: ID=%d, UserID=%d, Status=%s\n\n", taskOutput.ID, taskOutput.UserID, taskOutput.Status)

	// 示例2: 触发发布任务
	fmt.Println("--- Example 2: Trigger Publish Task ---")
	publishEvent := &dto.PublishEventDTO{
		UserID:       12345,
		ContentID:    999,
		TopicIDs:     []uint64{1001, 1003},
		LikeCount:    15,
		CommentCount: 5,
		IsAudited:    true,
		AuditStatus:  1,
	}

	triggerInput := dto.TriggerTaskInput{
		TaskMode: publishEvent,
	}

	if err := container.TriggerTaskUC.Execute(ctx, triggerInput); err != nil {
		log.Printf("Trigger task failed: %v", err)
	}
	fmt.Println()

	// 示例3: 创建签到任务
	fmt.Println("--- Example 3: Create Checkin Task ---")
	checkinCreateInput := dto.CreateTaskInput{
		ActivityID:   2,
		TaskID:       200,
		UserID:       12345,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	checkinTaskOutput, err := container.CreateTaskUC.Execute(ctx, checkinCreateInput)
	if err != nil {
		log.Fatalf("Create checkin task failed: %v", err)
	}
	fmt.Printf("Checkin task created: ID=%d\n\n", checkinTaskOutput.ID)

	// 示例4: 触发签到任务
	fmt.Println("--- Example 4: Trigger Checkin Task ---")
	checkinEvent := &dto.CheckinEventDTO{
		UserID: 12345,
		Date:   "2024-01-01",
	}

	checkinTriggerInput := dto.TriggerTaskInput{
		TaskMode: checkinEvent,
	}

	if err := container.TriggerTaskUC.Execute(ctx, checkinTriggerInput); err != nil {
		log.Printf("Trigger checkin task failed: %v", err)
	}
	fmt.Println()

	// 示例5: 查询用户任务列表
	fmt.Println("--- Example 5: Query User Tasks ---")
	tasks, err := container.QueryTaskUC.ExecuteList(ctx, 12345)
	if err != nil {
		log.Printf("Query tasks failed: %v", err)
	} else {
		fmt.Printf("Found %d tasks for user 12345:\n", len(tasks))
		for i, task := range tasks {
			fmt.Printf("  %d. Task ID=%d, Status=%s, Progress=%d/%d\n",
				i+1, task.ID, task.Status, task.Progress, task.Target)
		}
	}
	fmt.Println()

	// 示例6: 活动管理
	fmt.Println("--- Example 6: Activity Management ---")
	activity := &entity.ActActivity{
		Name:      "Spring Festival Activity",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:    entity.ActivityStatusActive,
	}
	if err := container.ActivityRepo.Create(ctx, activity); err != nil {
		log.Printf("Create activity failed: %v", err)
	} else {
		fmt.Printf("Activity created: ID=%d, Name=%s, IsActive=%v\n",
			activity.ID, activity.Name, activity.IsActive())
	}

	fmt.Println("\n=== Example execution completed ===")
	container.Logger.Info("Application finished successfully")
}

