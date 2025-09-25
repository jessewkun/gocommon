# Cron 定时任务模块

## 简介

`cron` 模块提供完整的定时任务管理功能，支持任务注册、调度执行、手动触发、超时控制、钩子机制等特性。基于 `robfig/cron/v3` 实现，提供企业级的定时任务解决方案。

## 主要特性

- **🕐 灵活调度**：支持标准 cron 表达式，精确控制任务执行时间
- **🛡️ 安全执行**：集成 `safego` 保护，防止 panic 导致服务崩溃
- **⏱️ 超时控制**：支持任务超时设置，自动处理超时任务
- **🔗 钩子机制**：提供 `BeforeRun` 和 `AfterRun` 钩子，支持任务前置和后置处理
- **🎯 手动触发**：支持手动执行指定任务，便于测试和运维
- **📊 完整日志**：详细的执行日志，包含链路追踪和性能统计
- **⚙️ 全局管理**：支持全局任务管理器，简化任务注册和管理
- **🚫 任务隔离**：单个任务异常不影响其他任务执行

## 快速开始

### 1. 基本使用

```go
package main

import (
    "context"
    "time"

    "github.com/jessewkun/gocommon/cron"
)

// 定义任务
type MyTask struct {
    cron.BaseTask
}

func (t *MyTask) Run(ctx context.Context) error {
    // 业务逻辑
    return nil
}

func main() {
    // 创建管理器
    manager := cron.NewManager()

    // 注册任务
    manager.RegisterTask(&MyTask{
        BaseTask: cron.BaseTask{
            TaskName:    "my_task",
            TaskDesc:    "我的定时任务",
            TaskEnabled: true,
            TaskSpec:    "0 */5 * * * *", // 每5分钟执行一次
            TaskTimeout: 30 * time.Second,
        },
    })

    // 启动管理器
    ctx := context.Background()
    manager.Start(ctx)

    // 程序退出时停止管理器
    defer manager.Stop()
}
```

### 2. 自动注册（推荐）

```go
package main

import (
    "context"
    "time"

    "github.com/jessewkun/gocommon/cron"
)

type DataCleanupTask struct {
    cron.BaseTask
}

func (t *DataCleanupTask) Run(ctx context.Context) error {
    // 数据清理逻辑
    return nil
}

// 自动注册任务
func init() {
    cron.AutoRegisterTask(&DataCleanupTask{
        BaseTask: cron.BaseTask{
            TaskName:    "data_cleanup",
            TaskDesc:    "数据清理任务",
            TaskEnabled: true,
            TaskSpec:    "0 0 2 * * *", // 每天凌晨2点执行
            TaskTimeout: 10 * time.Minute,
        },
    })
}

func main() {
    // 创建管理器会自动注册所有任务
    manager := cron.NewManager()

    // 启动管理器
    ctx := context.Background()
    manager.Start(ctx)
    defer manager.Stop()
}
```

## 详细功能

### 任务配置

#### BaseTask 结构

```go
type BaseTask struct {
    TaskName    string        // 任务名称（唯一标识）
    TaskDesc    string        // 任务描述
    TaskEnabled bool          // 是否启用任务
    TaskSpec    string        // cron 调度表达式
    TaskTimeout time.Duration // 任务超时时间（0表示不超时）
}
```

#### Cron 表达式格式

支持标准 cron 表达式，包含秒级精度：

```
格式：秒 分 时 日 月 周
示例：
- "0 * * * * *"     // 每分钟执行
- "0 */5 * * * *"   // 每5分钟执行
- "0 0 */2 * * *"   // 每2小时执行
- "0 0 0 * * *"     // 每天午夜执行
- "0 0 9 * * 1"     // 每周一上午9点执行
```

### 钩子机制

#### BeforeRun - 任务执行前

```go
func (t *MyTask) BeforeRun(ctx context.Context) error {
    // 检查依赖服务
    if !db.IsConnected() {
        return fmt.Errorf("database not connected")
    }

    // 初始化资源
    t.initResources()

    // 记录任务开始
    logger.Info(ctx, "TASK", "Task %s starting", t.TaskName)

    return nil
}
```

#### AfterRun - 任务执行后

```go
func (t *MyTask) AfterRun(ctx context.Context) error {
    // 清理资源
    t.cleanupResources()

    // 发送通知
    t.sendNotification()

    // 更新状态
    t.updateTaskStatus()

    return nil
}
```

### 手动执行任务

```go
ctx := context.Background()
manager := cron.NewManager()
err := manager.RunTask(ctx, "my_task")
if err != nil {
    log.Printf("任务执行失败: %v", err)
}
```

### 超时控制

```go
type LongRunningTask struct {
    cron.BaseTask
}

func (t *LongRunningTask) Run(ctx context.Context) error {
    // 长时间运行的业务逻辑
    // 框架会自动处理超时，业务代码无需检查 ctx.Done()

    for i := 0; i < 1000; i++ {
        // 业务处理
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}

// 配置超时时间
AutoRegisterTask(&LongRunningTask{
    BaseTask: cron.BaseTask{
        TaskName:    "long_task",
        TaskEnabled: true,
        TaskSpec:    "0 */10 * * * *",
        TaskTimeout: 5 * time.Minute, // 5分钟超时
    },
})
```

### 错误处理

#### 任务级别错误处理

```go
func (t *MyTask) Run(ctx context.Context) error {
    // 业务逻辑可能返回错误
    if err := t.processData(); err != nil {
        // 返回错误会被框架记录和处理
        return fmt.Errorf("数据处理失败: %w", err)
    }

    return nil
}
```

#### Panic 保护

```go
func (t *MyTask) Run(ctx context.Context) error {
    // 即使这里发生 panic，框架也会安全捕获
    panic("模拟业务异常")

    return nil
}
```

## 高级用法

### 多管理器场景

```go
// 创建不同的管理器处理不同类型的任务
userManager := cron.NewManager()
orderManager := cron.NewManager()

// 注册用户相关任务
userManager.RegisterTask(&UserTask{...})

// 注册订单相关任务
orderManager.RegisterTask(&OrderTask{...})

// 分别启动
userManager.Start(ctx)
orderManager.Start(ctx)
```

### 任务状态管理

```go
type StatusTask struct {
    cron.BaseTask
    status string
}

func (t *StatusTask) BeforeRun(ctx context.Context) error {
    t.status = "running"
    t.saveStatus()
    return nil
}

func (t *StatusTask) AfterRun(ctx context.Context) error {
    t.status = "completed"
    t.saveStatus()
    return nil
}

func (t *StatusTask) Run(ctx context.Context) error {
    // 业务逻辑
    return nil
}
```

## 最佳实践

### 1. 任务设计原则

- **单一职责**：每个任务只做一件事
- **幂等性**：任务可以重复执行而不产生副作用
- **可恢复性**：任务失败后可以重新执行
- **监控友好**：提供足够的日志和状态信息

### 2. 资源管理

```go
func (t *MyTask) BeforeRun(ctx context.Context) error {
    // 初始化资源
    t.db = db.NewConnection()
    t.redis = redis.NewClient()
    return nil
}

func (t *MyTask) AfterRun(ctx context.Context) error {
    // 清理资源
    if t.db != nil {
        t.db.Close()
    }
    if t.redis != nil {
        t.redis.Close()
    }
    return nil
}
```

### 3. 性能优化

- 合理设置超时时间，避免任务长时间阻塞
- 使用异步执行处理耗时任务
- 避免在任务中执行阻塞操作
- 合理使用 BeforeRun/AfterRun 钩子

## API 参考

### 核心接口

```go
// Task 定时任务接口
type Task interface {
    Name() string                                    // 任务名称
    Spec() string                                    // cron 表达式
    BeforeRun(ctx context.Context) error            // 执行前钩子
    Run(ctx context.Context) error                   // 任务执行
    AfterRun(ctx context.Context) error             // 执行后钩子
    Timeout() time.Duration                          // 超时时间
    Enabled() bool                                   // 是否启用
}
```

### 管理器方法

```go
// 创建管理器
func NewManager() *Manager

// 注册任务
func (m *Manager) RegisterTask(task Task) error

// 启动管理器
func (m *Manager) Start(ctx context.Context) error

// 停止管理器
func (m *Manager) Stop()

// 手动执行任务
func (m *Manager) RunTask(ctx context.Context, taskName string) error

// 获取任务名称列表
func (m *Manager) GetTaskNames() []string

// 检查是否运行中
func (m *Manager) IsRunning() bool
```

### 全局函数

```go
// 自动注册任务
func AutoRegisterTask(task Task) error
```

## 注意事项

1. **任务名称唯一性**：确保每个任务名称在管理器中唯一
2. **Cron 表达式格式**：使用 6 位格式（包含秒）
3. **超时设置**：合理设置任务超时时间，避免资源占用
4. **错误处理**：在任务中正确处理错误，避免静默失败
5. **资源清理**：在 AfterRun 中清理资源，防止内存泄漏
6. **并发安全**：任务本身不需要考虑并发安全，框架保证串行执行

## 示例项目

完整的使用示例请参考 [example.go](./example.go) 文件。
