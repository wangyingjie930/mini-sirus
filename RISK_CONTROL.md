# 风控系统实现文档

## 概述

本项目实现了一个完整的风控检查系统，用于监控和防范用户的异常行为。风控系统作为**用例层的同步依赖**，在任务完成**之前**执行风控检查，确保只有通过风控的任务才能被保存和奖励。

## 架构设计

### 1. 分层架构

```
用例层 (UseCase)
    ↓
    ├─ 风控服务 (同步依赖，阻塞任务完成)
    └─ 观察者 (异步通知，非阻塞操作)
    ↓
输出端口 (Output Port)
    ↓
适配器层 (Adapter)
    ↓
仓储实现 (Repository)
```

### 2. 风控服务的正确使用方式

#### ❌ 错误的做法：作为观察者

```go
// 错误：风控作为观察者，在任务已完成后异步执行
type RiskCheckObserver struct {
    riskCheckService output.RiskCheckService
}

func (o *RiskCheckObserver) OnTaskDetailCreated(ctx context.Context, detail *entity.ActUserTaskDetail) error {
    // 此时任务已经保存，即使风控失败也无法回滚
    return o.riskCheckService.Check(ctx, detail.UserID)
}
```

**问题**：
- 任务已经入库，风控失败无法阻止
- 奖励可能已经发放，事后补救成本高
- 违反业务逻辑：应该先检查，再完成

#### ✅ 正确的做法：作为用例依赖

```go
// 正确：风控服务作为用例的依赖注入
type TriggerTaskUseCase struct {
    riskCheckService output.RiskCheckService // 依赖注入
}

func (uc *TriggerTaskUseCase) processTask(ctx context.Context, task *entity.ActUserTask) error {
    // 1. 先执行风控检查（同步阻塞）
    if err := uc.performRiskCheck(ctx, task.UserID, task.ID); err != nil {
        return fmt.Errorf("风控检查失败: %w", err)
    }
    
    // 2. 风控通过后才创建任务明细
    detail := &entity.ActUserTaskDetail{...}
    
    // 3. 保存任务
    if err := uc.taskDetailRepo.Create(ctx, detail); err != nil {
        return err
    }
    
    // 4. 通知观察者（触达、统计等非阻塞操作）
    uc.observerRegistry.Notify(ctx, detail)
    
    return nil
}
```

**优势**：
- 风控检查在任务完成前执行，可以阻止恶意行为
- 失败时不产生脏数据
- 符合业务逻辑顺序

### 3. 观察者模式的正确使用

观察者模式适用于：
- ✅ 触达通知（推送消息）
- ✅ 异步统计（用户行为分析）
- ✅ 日志记录
- ✅ 积分发放（可补偿）

观察者模式**不适用**于：
- ❌ 风控检查（必须阻塞）
- ❌ 权限验证（必须阻塞）
- ❌ 余额扣减（不可补偿）
- ❌ 任何需要阻止业务流程的操作

### 3. 核心组件

#### 2.1 输出端口接口

**文件**: `internal/usecase/port/output/risk_check_service.go`

定义了风控服务的抽象接口：

- `CheckUserBehavior`: 检查用户行为异常
- `CheckTaskFrequency`: 检查任务完成频率
- `CheckDeviceFingerprint`: 检查设备指纹
- `RecordTaskCompletion`: 记录任务完成事件
- `IsUserBlacklisted`: 检查用户是否在黑名单
- `AddToBlacklist`: 将用户加入黑名单

#### 2.2 风控服务实现

**文件**: `internal/adapter/repository/memory/risk_check_repository.go`

内存实现的风控服务，包含以下核心功能：

##### 用户行为检查
- 检测短时间内的大量操作（1分钟内 > 10次）
- 检测操作时间间隔的规律性（机器人特征）
- 使用方差分析识别自动化脚本

##### 任务频率检查
- 1小时内同一任务完成次数限制（< 10次）
- 24小时内所有任务完成次数限制（< 100次）
- 新用户异常活跃检测（24小时内 > 20次）

##### 设备指纹检查
- 单设备关联账号数量限制（≤ 5个）
- 单用户使用设备数量限制（≤ 10个）
- 双向映射关系维护

#### 2.3 风控检查流程

**文件**: `internal/usecase/task/trigger_task.go`

在 `TriggerTaskUseCase` 中，风控检查在任务明细创建**之前**同步执行：

```go
// processTask 处理单个任务
func (uc *TriggerTaskUseCase) processTask(ctx context.Context, task *entity.ActUserTask) error {
    // 1. 执行规则引擎判定
    reach, err := uc.ruleEngine.Evaluate(...)
    if !reach {
        return nil
    }

    // 2. 风控检查（同步执行，阻塞任务完成）⭐ 关键步骤
    if err := uc.performRiskCheck(ctx, task.UserID, task.ID); err != nil {
        return fmt.Errorf("风控检查失败: %w", err)
    }

    // 3. 风控通过后，才创建任务明细
    detail := &entity.ActUserTaskDetail{...}
    if err := uc.taskDetailRepo.Create(ctx, detail); err != nil {
        return err
    }

    // 4. 更新任务进度
    task.UpdateProgress()
    uc.taskRepo.Update(ctx, task)

    // 5. 记录任务完成事件（用于风控统计）
    uc.riskCheckService.RecordTaskCompletion(ctx, task.UserID, task.ID, detail.CreatedAt)

    // 6. 通知观察者（触达服务、统计服务等非阻塞操作）
    uc.observerRegistry.Notify(ctx, detail)

    return nil
}

// performRiskCheck 执行风控检查（同步阻塞）
func (uc *TriggerTaskUseCase) performRiskCheck(ctx context.Context, userID, taskID int64) error {
    // 检查黑名单
    if isBlacklisted, _ := uc.riskCheckService.IsUserBlacklisted(ctx, userID); isBlacklisted {
        return fmt.Errorf("用户已被列入黑名单")
    }

    // 检查用户行为
    if err := uc.riskCheckService.CheckUserBehavior(ctx, userID, nil); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "用户行为异常")
        return err
    }

    // 检查任务频率
    if err := uc.riskCheckService.CheckTaskFrequency(ctx, userID, taskID); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "任务完成频率过高")
        return err
    }

    // 检查设备指纹
    if err := uc.riskCheckService.CheckDeviceFingerprint(ctx, userID, nil); err != nil {
        _ = uc.riskCheckService.AddToBlacklist(ctx, userID, "设备指纹异常")
        return err
    }

    return nil
}
```

#### 2.4 触达服务（观察者模式）

**文件**: `internal/adapter/observer/task_observers.go`

触达服务适合作为观察者，因为它是**非阻塞**的通知操作：

## 风控规则详解

### 1. 用户行为异常检测

#### 规则1: 操作频率检测
```go
// 1分钟内操作超过10次 → 异常
if recentCount > 10 {
    return fmt.Errorf("用户行为异常: 1分钟内操作次数过多(%d次)", recentCount)
}
```

**适用场景**: 防止机器人或脚本快速刷任务

#### 规则2: 时间间隔规律性检测
```go
// 操作时间间隔方差 < 0.1秒 → 过于规律（机器人特征）
if variance < 0.1 {
    return fmt.Errorf("用户行为异常: 操作时间间隔过于规律")
}
```

**适用场景**: 识别自动化脚本（人类操作通常有随机性）

### 2. 任务完成频率限制

#### 规则3: 单任务频率限制
```go
// 1小时内完成同一任务超过10次 → 异常
if taskCount >= 10 {
    return fmt.Errorf("任务完成频率过高")
}
```

**适用场景**: 防止单个任务被重复刷

#### 规则4: 总任务频率限制
```go
// 24小时内完成超过100个任务 → 异常
if totalCount >= 100 {
    return fmt.Errorf("任务完成频率过高")
}
```

**适用场景**: 防止羊毛党批量刷任务

#### 规则5: 新用户异常活跃
```go
// 新用户24小时内完成超过20个任务 → 异常
if len(completions) < 50 && totalCount > 20 {
    return fmt.Errorf("新用户异常活跃")
}
```

**适用场景**: 识别专门用于刷任务的小号

### 3. 设备指纹检测

#### 规则6: 一机多号
```go
// 单设备关联账号超过5个 → 异常
if accountCount > 5 {
    return fmt.Errorf("设备指纹异常: 单设备关联账号过多")
}
```

**适用场景**: 检测使用同一设备注册多个账号的行为

#### 规则7: 频繁换设备
```go
// 单用户使用设备超过10个 → 异常
if deviceCount > 10 {
    return fmt.Errorf("设备指纹异常: 用户使用设备过多")
}
```

**适用场景**: 检测异常的设备切换行为

### 4. 黑名单机制

一旦检测到风险行为，系统会自动将用户加入黑名单：

```go
// 加入黑名单
_ = o.riskCheckService.AddToBlacklist(ctx, detail.UserID, "用户行为异常")
```

黑名单用户的后续操作将被直接拒绝：

```go
if isBlacklisted {
    return fmt.Errorf("用户已被列入黑名单，禁止完成任务")
}
```

## 数据结构

### 用户行为记录
```go
type UserBehaviorRecord struct {
    UserID    int64
    Action    string
    TaskID    int64
    Timestamp time.Time
}
```

### 任务完成记录
```go
type TaskCompletionRecord struct {
    UserID    int64
    TaskID    int64
    Timestamp time.Time
}
```

## 使用示例

### 1. 初始化风控服务（依赖注入）

```go
// 创建风控服务
riskCheckService := memory.NewRiskCheckServiceMemory()

// 将风控服务注入到用例中（不是观察者！）
triggerTaskUC := task.NewTriggerTaskUseCase(
    taskRepo,
    taskDetailRepo,
    ruleEngine,
    observerRegistry,
    distributedLock,
    riskCheckService, // 作为依赖注入
)

// 注册观察者（仅注册适合异步执行的观察者）
checkinObserver := observer.NewCheckinReachObserver(reachAdapter)
observerRegistry.Register(checkinObserver)
// 注意：不再注册 RiskCheckObserver
```

### 2. 自动触发

风控检查会在任务完成**之前**自动执行，失败时会阻止任务完成：

```go
// 触发任务
err := container.TriggerTaskUC.Execute(ctx, triggerInput)
// 执行流程：
// 1. 规则引擎判定 → 
// 2. 风控检查（同步阻塞）→ 
// 3. 创建任务明细 → 
// 4. 更新进度 → 
// 5. 通知观察者（异步）

if err != nil {
    // 风控失败，任务未完成，没有产生脏数据
    log.Printf("任务触发失败: %v", err)
}
```

### 3. 手动检查

也可以直接调用风控服务进行检查：

```go
// 检查用户是否在黑名单
isBlacklisted, err := riskCheckService.IsUserBlacklisted(ctx, userID)

// 检查任务频率
err := riskCheckService.CheckTaskFrequency(ctx, userID, taskID)
```

## 测试用例

项目提供了完整的风控测试用例（`cmd/example/main.go`）：

1. **测试1**: 正常签到 - 验证正常用户可以通过风控检查
2. **测试2**: 频繁签到 - 模拟1分钟内12次签到，**风控会在任务保存前拦截**
3. **测试3**: 黑名单检查 - 验证用户被加入黑名单
4. **测试4**: 黑名单用户尝试操作 - 验证黑名单用户**在任务创建前被拒绝**

运行测试：

```bash
go run cmd/example/main.go
```

**重要**：新架构下，风控失败的任务**不会被创建**，不会产生脏数据。

## 扩展建议

### 1. 实际项目中的增强

当前实现是一个基础版本，真实项目中建议增加：

#### 高级检测
- IP地址风险检测（代理、机房IP、VPN）
- 地理位置跳变检测
- 生物特征分析（陀螺仪、触摸压感）
- 验证码人机识别

#### 机器学习
- 用户行为模型训练
- 异常检测算法（Isolation Forest、LSTM）
- 风险评分模型
- 黑产团伙识别

#### 数据持久化
- 使用 Redis 存储实时数据（行为记录、频率统计）
- 使用 MySQL/PostgreSQL 存储历史数据
- 使用 HBase/ES 存储大规模行为日志

#### 监控告警
- 实时风控指标监控
- 异常行为告警
- 风控规则效果评估
- A/B测试框架

### 2. 性能优化

- 使用滑动窗口算法优化时间范围查询
- 使用布隆过滤器加速黑名单检查
- 使用时间轮算法清理过期数据
- 分布式限流（令牌桶、漏桶算法）

### 3. 规则引擎

建议将风控规则配置化：

```yaml
rules:
  - name: "频率检查"
    type: "frequency"
    window: "1h"
    threshold: 10
    action: "block"
  
  - name: "新用户活跃"
    type: "new_user_activity"
    window: "24h"
    threshold: 20
    action: "review"
```

## 总结

本风控系统实现了：

✅ 完整的分层架构（端口-适配器模式）  
✅ **风控服务作为用例依赖**（同步阻塞，在任务完成前执行）  
✅ 观察者模式用于非阻塞操作（触达、统计等）  
✅ 多维度风控检测  
✅ 黑名单机制  
✅ 内存存储实现  
✅ 完整的测试用例  

### 关键设计原则

1. **风控必须同步**：在任务完成前执行，失败可阻止任务创建
2. **观察者用于通知**：用于触达、统计等不影响业务流程的操作
3. **职责分离清晰**：阻塞操作在用例层，非阻塞操作用观察者
4. **无脏数据产生**：风控失败时，任务不会入库

适用于中小型项目的风控需求，可根据实际业务场景进行扩展和优化。

