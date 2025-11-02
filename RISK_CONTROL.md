# 风控系统实现文档

## 概述

本项目实现了一个完整的风控检查系统，用于监控和防范用户的异常行为。风控系统作为任务系统的观察者（Observer），在任务完成时自动触发风控检查。

## 架构设计

### 1. 分层架构

```
用例层 (UseCase)
    ↓
输出端口 (Output Port)
    ↓
适配器层 (Adapter)
    ↓
仓储实现 (Repository)
```

### 2. 核心组件

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

#### 2.3 风控观察者

**文件**: `internal/adapter/observer/task_observers.go`

风控检查观察者 `RiskCheckObserver`，在任务明细创建时自动执行风控检查：

1. 检查用户是否在黑名单
2. 检查用户行为异常
3. 检查任务完成频率
4. 检查设备指纹
5. 记录任务完成事件

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

### 1. 初始化风控服务

```go
// 创建风控服务
riskCheckService := memory.NewRiskCheckServiceMemory()

// 创建风控观察者
riskCheckObserver := observer.NewRiskCheckObserver(riskCheckService)

// 注册观察者
observerRegistry.Register(riskCheckObserver)
```

### 2. 自动触发

风控检查会在任务完成时自动触发，无需手动调用：

```go
// 触发任务
container.TriggerTaskUC.Execute(ctx, triggerInput)
// → 任务完成 → 观察者通知 → 风控检查自动执行
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
2. **测试2**: 频繁签到 - 模拟1分钟内12次签到，触发风控拦截
3. **测试3**: 黑名单检查 - 验证用户被加入黑名单
4. **测试4**: 黑名单用户尝试操作 - 验证黑名单用户被拒绝

运行测试：

```bash
go run cmd/example/main.go
```

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
✅ 观察者模式集成  
✅ 多维度风控检测  
✅ 黑名单机制  
✅ 内存存储实现  
✅ 完整的测试用例  

适用于中小型项目的风控需求，可根据实际业务场景进行扩展和优化。

