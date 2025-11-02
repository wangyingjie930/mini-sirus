# Mini-Sirus - 任务系统

基于 **Clean Architecture（简洁架构）** 设计的任务系统，用于管理活动任务、用户进度、奖励发放等业务场景。

## 🏗️ 架构设计

本项目采用简洁架构（Clean Architecture）设计，遵循依赖倒置原则，确保业务逻辑独立于框架和基础设施。

```
┌─────────────────────────────────────────┐
│         Interface Layer                 │  HTTP/gRPC/Event Handlers
│         (外部接口层)                     │
└────────────────┬────────────────────────┘
                 │ depends on
                 ↓
┌─────────────────────────────────────────┐
│      Adapter Layer (适配器层)           │  Repository/RuleEngine/Observer
│      实现输出端口                        │
└────────────────┬────────────────────────┘
                 │ implements
                 ↓
┌─────────────────────────────────────────┐
│      Use Case Layer (用例层)            │  业务流程编排
│      定义输入/输出端口                   │
└────────────────┬────────────────────────┘
                 │ depends on
                 ↓
┌─────────────────────────────────────────┐
│      Domain Layer (领域层)              │  实体、值对象、领域事件
│      核心业务规则                        │
└─────────────────────────────────────────┘
```

## 📁 目录结构

```
mini-sirus/
├── cmd/
│   └── example/              # 示例程序入口
│       └── main.go
│
├── internal/
│   ├── domain/               # 领域层 - 核心业务逻辑
│   │   ├── entity/           # 实体对象
│   │   ├── valueobject/      # 值对象
│   │   ├── event/            # 领域事件
│   │   └── repository/       # 仓储接口定义
│   │
│   ├── usecase/              # 用例层 - 业务流程编排
│   │   ├── task/             # 任务相关用例
│   │   ├── dto/              # 数据传输对象
│   │   └── port/             # 端口定义
│   │       ├── input/        # 输入端口（服务接口）
│   │       └── output/       # 输出端口（依赖接口）
│   │
│   ├── adapter/              # 适配器层 - 实现输出端口
│   │   ├── repository/       # 仓储实现
│   │   │   └── memory/       # 内存实现
│   │   ├── rule_engine/      # 规则引擎适配器
│   │   ├── observer/         # 观察者实现
│   │   └── notification/     # 通知服务适配器
│   │
│   ├── infrastructure/       # 基础设施层
│   │   ├── config/           # 配置管理
│   │   ├── logger/           # 日志
│   │   └── lock/             # 分布式锁
│   │
│   └── interface/            # 接口层 - 外部访问入口
│       └── http/             # HTTP接口
│           ├── handler/      # 处理器
│           └── router/       # 路由
│
├── docs/
│   └── CLEAN_ARCHITECTURE.md # 架构详细文档
│
├── go.mod
├── go.sum
└── README.md
```

## Q&A

- createtask: 注册的待完成的任务 
- triggertask: 当触发某类事件时看到的任务


> 假如活动刚开始, 用户点击链接跳转登录时, 马上就注册它在这个活动的所有代办任务, 是否可行?
- https://gemini.google.com/share/5a7d568b1f49
- 大型活动（所有人都一样） -> 模式一：异步批量预创建。
- 事件型任务（如“发布文章”） -> 模式二：实时延迟创建。