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
	RiskCheckService *memory.RiskCheckServiceMemory

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
	riskCheckService := memory.NewRiskCheckServiceMemory()

	// 注册观察者
	checkinObserver := observer.NewCheckinReachObserver(reachAdapter)
	riskCheckObserver := observer.NewRiskCheckObserver(riskCheckService)
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
		RiskCheckService: riskCheckService,
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

	// 示例7: 风控测试 - 模拟频繁操作
	fmt.Println("\n--- Example 7: Risk Control Test ---")
	testRiskControl(ctx, container)
}

// testRiskControl 测试风控功能
func testRiskControl(ctx context.Context, container *Container) {
	// 创建测试用户的签到任务
	fmt.Println("创建测试用户(999)的签到任务...")
	testCreateInput := dto.CreateTaskInput{
		ActivityID:   3,
		TaskID:       300,
		UserID:       999,
		Target:       1,
		TaskType:     valueobject.TaskTypeCheckin,
		TaskCondExpr: "IS_TODAY()",
	}

	testTask, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
	if err != nil {
		log.Printf("创建测试任务失败: %v", err)
		return
	}
	fmt.Printf("测试任务创建成功: ID=%d\n\n", testTask.ID)

	// 测试1: 正常签到
	fmt.Println("测试1: 正常签到...")
	normalCheckin := &dto.CheckinEventDTO{
		UserID: 999,
		Date:   time.Now().Format("2006-01-02"),
	}
	triggerInput := dto.TriggerTaskInput{
		TaskMode: normalCheckin,
	}

	if err := container.TriggerTaskUC.Execute(ctx, triggerInput); err != nil {
		log.Printf("❌ 正常签到失败: %v", err)
	} else {
		fmt.Println("✅ 正常签到成功，风控检查通过")
	}

	time.Sleep(100 * time.Millisecond)

	// 测试2: 模拟短时间内频繁签到（触发风控）
	fmt.Println("\n测试2: 模拟短时间内频繁签到...")
	for i := 0; i < 12; i++ {
		// 重新创建任务（因为之前的已完成）
		testCreateInput.TaskID = int64(300 + i + 1)
		_, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
		if err != nil {
			continue
		}

		checkinEvent := &dto.CheckinEventDTO{
			UserID: 999,
			Date:   time.Now().Format("2006-01-02"),
		}
		triggerInput := dto.TriggerTaskInput{
			TaskMode: checkinEvent,
		}

		if err := container.TriggerTaskUC.Execute(ctx, triggerInput); err != nil {
			fmt.Printf("❌ 第%d次签到被风控拦截: %v\n", i+1, err)
			break
		} else {
			fmt.Printf("✅ 第%d次签到成功\n", i+1)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// 测试3: 检查用户是否被加入黑名单
	fmt.Println("\n测试3: 检查用户是否被加入黑名单...")
	isBlacklisted, err := container.RiskCheckService.IsUserBlacklisted(ctx, 999)
	if err != nil {
		log.Printf("检查黑名单失败: %v", err)
	} else if isBlacklisted {
		fmt.Println("⚠️  用户999已被加入黑名单")
	} else {
		fmt.Println("用户999未在黑名单中")
	}

	// 测试4: 尝试黑名单用户签到
	if isBlacklisted {
		fmt.Println("\n测试4: 黑名单用户尝试签到...")
		testCreateInput.TaskID = 400
		_, err := container.CreateTaskUC.Execute(ctx, testCreateInput)
		if err != nil {
			log.Printf("创建任务失败: %v", err)
			return
		}

		blacklistCheckin := &dto.CheckinEventDTO{
			UserID: 999,
			Date:   time.Now().Format("2006-01-02"),
		}
		triggerInput := dto.TriggerTaskInput{
			TaskMode: blacklistCheckin,
		}

		if err := container.TriggerTaskUC.Execute(ctx, triggerInput); err != nil {
			fmt.Printf("❌ 黑名单用户签到被拒绝: %v\n", err)
		} else {
			fmt.Println("⚠️  黑名单用户签到成功（不应该发生）")
		}
	}

	fmt.Println("\n风控测试完成！")
}

