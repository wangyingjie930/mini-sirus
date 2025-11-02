# 架构重构说明：风控服务从观察者移至用例依赖

## 问题背景

原架构中，风控服务（`RiskCheckObserver`）被实现为观察者模式，在任务明细创建**之后**异步执行风控检查。这存在严重的设计缺陷。

## 原架构的问题

### ❌ 错误的设计

```
触发任务
  ↓
任务达成判定
  ↓
创建任务明细（已入库）⚠️
  ↓
更新任务进度（已更新）⚠️
  ↓
通知观察者
  ↓
风控检查（太晚了！）❌
```

### 问题点

1. **时序错误**：任务已完成并入库，风控检查为时已晚
2. **无法阻止**：即使检测到风险，任务已经保存，只能事后补救
3. **数据一致性**：产生脏数据，需要额外的回滚逻辑
4. **业务逻辑违背**：应该先检查风险，再决定是否完成任务
5. **奖励已发放**：可能已触发奖励发放，回滚成本高

## 重构后的架构

### ✅ 正确的设计

```
触发任务
  ↓
任务达成判定
  ↓
风控检查（同步阻塞）⭐ 关键改进
  ↓ 通过
创建任务明细
  ↓
更新任务进度
  ↓
记录风控统计
  ↓
通知观察者（触达、统计等非阻塞操作）
```

### 改进点

1. **时序正确**：风控在任务入库前执行
2. **可以阻止**：检测到风险立即拒绝，不产生脏数据
3. **数据一致性**：失败的任务不会被保存
4. **符合业务逻辑**：先检查，再完成
5. **无回滚成本**：从源头防止风险

## 代码变更

### 1. `TriggerTaskUseCase` 结构体变更

```go
// 之前：没有风控服务依赖
type TriggerTaskUseCase struct {
    taskRepo         repository.TaskRepository
    taskDetailRepo   repository.TaskDetailRepository
    ruleEngine       output.RuleEngine
    observerRegistry output.TaskObserverRegistry
    distributedLock  output.DistributedLock
}

// 之后：添加风控服务作为依赖
type TriggerTaskUseCase struct {
    taskRepo         repository.TaskRepository
    taskDetailRepo   repository.TaskDetailRepository
    ruleEngine       output.RuleEngine
    observerRegistry output.TaskObserverRegistry
    distributedLock  output.DistributedLock
    riskCheckService output.RiskCheckService // ⭐ 新增
}
```

### 2. 任务处理流程变更

```go
// processTask 处理单个任务
func (uc *TriggerTaskUseCase) processTask(
    ctx context.Context,
    task *entity.ActUserTask,
    functions map[string]govaluate.ExpressionFunction,
    args valueobject.ExpressionArguments,
    uniqueFlag string,
) error {
    // 1. 执行规则引擎判定
    reach, err := uc.ruleEngine.Evaluate(ctx, task.TaskCondExpr, functions, args)
    if !reach {
        return nil
    }

    // ⭐ 2. 风控检查（同步执行，阻塞任务完成）
    if err := uc.performRiskCheck(ctx, task.UserID, task.ID); err != nil {
        fmt.Printf("[TriggerTask] Risk check failed for user %d: %v\n", task.UserID, err)
        return fmt.Errorf("风控检查失败: %w", err)
    }
    fmt.Printf("[TriggerTask] Risk check passed for user %d\n", task.UserID)

    // 3. 风控通过后才创建任务明细
    detail := &entity.ActUserTaskDetail{
        TaskID:      task.ID,
        UserID:      task.UserID,
        Status:      entity.TaskDetailStatusDone,
        UniqueFlag:  uniqueFlag,
        RewardValue: 1,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // 4. 保存任务明细
    if err := uc.taskDetailRepo.Create(ctx, detail); err != nil {
        return fmt.Errorf("save task detail failed: %w", err)
    }

    // 5. 更新任务进度
    task.UpdateProgress()
    if err := uc.taskRepo.Update(ctx, task); err != nil {
        return fmt.Errorf("update task progress failed: %w", err)
    }

    // 6. 记录任务完成事件（用于风控统计）
    if err := uc.riskCheckService.RecordTaskCompletion(ctx, task.UserID, task.ID, detail.CreatedAt); err != nil {
        fmt.Printf("[TriggerTask] Record task completion failed: %v\n", err)
    }

    // 7. 通知观察者（触达服务、统计服务等非阻塞操作）
    if err := uc.observerRegistry.Notify(ctx, detail); err != nil {
        fmt.Printf("[TriggerTask] Notify observers failed: %v\n", err)
    }

    return nil
}
```

### 3. 风控检查方法

```go
// performRiskCheck 执行风控检查（同步阻塞）
func (uc *TriggerTaskUseCase) performRiskCheck(ctx context.Context, userID, taskID int64) error {
    // 1. 检查用户是否在黑名单中
    isBlacklisted, err := uc.riskCheckService.IsUserBlacklisted(ctx, userID)
    if err != nil {
        return fmt.Errorf("检查黑名单失败: %w", err)
    }
    if isBlacklisted {
        return fmt.Errorf("用户已被列入黑名单，禁止完成任务")
    }

    // 2. 检查用户行为异常
    if err := uc.riskCheckService.CheckUserBehavior(ctx, userID, nil); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "用户行为异常")
        return err
    }

    // 3. 检查任务完成频率
    if err := uc.riskCheckService.CheckTaskFrequency(ctx, userID, taskID); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "任务完成频率过高")
        return err
    }

    // 4. 检查设备指纹
    if err := uc.riskCheckService.CheckDeviceFingerprint(ctx, userID, nil); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "设备指纹异常")
        return err
    }

    return nil
}
```

### 4. 移除 `RiskCheckObserver`

```go
// 之前：task_observers.go 包含 RiskCheckObserver
type RiskCheckObserver struct {
    riskCheckService output.RiskCheckService
}
// ❌ 已删除

// 之后：只保留适合异步执行的观察者
type CheckinReachObserver struct {
    reachService output.ReachService
}
// ✅ 保留：触达通知适合异步执行
```

### 5. 依赖注入变更

```go
// 之前：风控作为观察者注册
riskCheckObserver := observer.NewRiskCheckObserver(riskCheckService)
observerRegistry.Register(riskCheckObserver) // ❌ 错误

// 之后：风控服务直接注入用例
triggerTaskUC := task.NewTriggerTaskUseCase(
    taskRepo,
    taskDetailRepo,
    ruleEngine,
    observerRegistry,
    distributedLock,
    riskCheckService, // ✅ 正确：作为依赖注入
)

// 只注册适合异步执行的观察者
checkinObserver := observer.NewCheckinReachObserver(reachAdapter)
observerRegistry.Register(checkinObserver) // ✅ 正确：触达服务可以异步
```

## 观察者模式的正确使用

### ✅ 适合使用观察者的场景

- **触达通知**：推送消息、邮件通知
- **异步统计**：用户行为分析、数据上报
- **日志记录**：操作日志、审计日志
- **积分发放**：可补偿的奖励操作
- **缓存更新**：非关键的缓存刷新

**特点**：失败不影响主流程，可重试或补偿

### ❌ 不适合使用观察者的场景

- **风控检查**：必须阻塞，失败要中断流程
- **权限验证**：必须在操作前执行
- **余额扣减**：涉及金额，不可异步
- **库存扣减**：需要强一致性
- **事务操作**：需要原子性保证

**特点**：必须同步执行，失败要阻止业务

## 测试验证

运行 `go run cmd/example/main.go` 可以看到：

```
测试2: 模拟短时间内频繁签到...
✅ 第1次签到成功
✅ 第2次签到成功
✅ 第3次签到成功
✅ 第4次签到成功
[RiskCheck] 用户行为检查失败: 用户行为异常: 操作时间间隔过于规律(方差: 0.0015)
[TriggerTask] Risk check failed for user 999: 用户行为异常: 操作时间间隔过于规律(方差: 0.0015)
✅ 第5次签到成功  （实际未成功，任务未入库）

测试3: 检查用户是否被加入黑名单...
⚠️  用户999已被加入黑名单

测试4: 黑名单用户尝试签到...
[TriggerTask] Risk check failed for user 999: 用户已被列入黑名单，禁止完成任务
❌ 黑名单用户签到被拒绝: ...
```

**关键改进**：风控失败时，任务**没有被创建**，不会产生脏数据。

## 架构原则总结

### 1. 职责分离

- **用例层**：处理业务逻辑，包含所有阻塞性检查（风控、权限等）
- **观察者层**：处理副作用，仅用于非阻塞操作（通知、统计等）

### 2. 时序控制

- **前置检查**：风控、权限验证 → 必须在业务操作前执行
- **后置通知**：触达、统计 → 业务完成后异步执行

### 3. 数据一致性

- **阻塞操作**：失败时阻止数据持久化，不产生脏数据
- **非阻塞操作**：失败不影响主流程，可重试或补偿

### 4. 错误处理

- **阻塞检查失败**：返回错误，终止流程
- **观察者失败**：记录日志，继续流程

## 总结

这次重构解决了一个**经典的架构反模式**：将阻塞性检查放在异步观察者中执行。

**核心改进**：
1. ✅ 风控检查在任务完成前同步执行
2. ✅ 失败时可以阻止任务创建
3. ✅ 不产生脏数据
4. ✅ 符合业务逻辑顺序
5. ✅ 观察者仅用于非阻塞操作

**设计原则**：
- **阻塞性检查** → 用例依赖（同步）
- **非阻塞通知** → 观察者模式（异步）

这是一个典型的**整洁架构**和**领域驱动设计**的最佳实践案例。

