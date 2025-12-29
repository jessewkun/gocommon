# 日志模块

## 功能简介

日志模块基于 [zap](https://github.com/uber-go/zap) 提供高性能、结构化的日志记录功能，支持多级别日志、日志轮转、报警集成等功能。

- ✅ 多级别日志：支持 Debug、Info、Warn、Error、Panic、Fatal 等日志级别
- ✅ 结构化日志：支持结构化字段记录，便于日志分析和查询
- ✅ 上下文传递：自动传递 trace_id、请求路径等上下文信息
- ✅ 日志轮转：基于 lumberjack 的日志轮转，支持按大小和时间轮转
- ✅ 报警集成：支持日志报警，异常情况自动发送告警
- ✅ 性能优化：零分配日志记录，高性能输出
- ✅ 系统信息：自动记录主机名、IP 等系统信息

## 配置说明

### Config 配置结构

```go
type Config struct {
    Path                 string   `mapstructure:"path" json:"path"`                                   // 日志文件路径
    Closed               bool     `mapstructure:"closed" json:"closed"`                               // 是否关闭日志
    MaxSize              int      `mapstructure:"max_size" json:"max_size"`                           // 单个日志文件最大大小（MB），默认 100MB
    MaxAge               int      `mapstructure:"max_age" json:"max_age"`                             // 日志文件最多保存天数，默认 30 天
    MaxBackup            int      `mapstructure:"max_backup" json:"max_backup"`                       // 保留的备份文件数量，默认 10 个
    TransparentParameter []string `mapstructure:"transparent_parameter" json:"transparent_parameter"` // 透传参数，从 context 中继承
    LogLevel             string   `mapstructure:"log_level" json:"log_level"`                         // 日志级别: debug, info, warn, error, fatal, panic
    AlarmLevel           string   `mapstructure:"alarm_level" json:"alarm_level"`                     // 报警级别: warn, error
}
```

### 配置示例

**TOML 格式：**

```toml
[log]
path = "./logs/app.log"
closed = false
max_size = 100        # 单个日志文件最大 100MB
max_age = 30          # 日志文件保存 30 天
max_backup = 10       # 保留 10 个备份文件
transparent_parameter = ["trace_id", "user_id"]
log_level = "info"    # 日志级别
alarm_level = "warn"  # 报警级别
```

**JSON 格式：**

```json
{
  "log": {
    "path": "./logs/app.log",
    "closed": false,
    "max_size": 100,
    "max_age": 30,
    "max_backup": 10,
    "transparent_parameter": ["trace_id", "user_id"],
    "log_level": "info",
    "alarm_level": "warn"
  }
}
```

### 配置建议

**高并发系统：**

```toml
[log]
max_size = 200    # 增大文件大小
max_age = 7       # 减小保存天数
max_backup = 20   # 增加备份数量
```

**低频率系统：**

```toml
[log]
max_size = 50     # 减小文件大小
max_age = 90      # 增大保存天数
max_backup = 5    # 减少备份数量
```

**磁盘空间紧张：**

```toml
[log]
max_size = 50
max_age = 7
max_backup = 3
```

## 基本使用

### 1. 初始化

日志模块会在配置加载时自动初始化，也可以手动调用：

```go
import "github.com/jessewkun/gocommon/logger"

if err := logger.Init(); err != nil {
    log.Fatalf("Failed to initialize logger: %v", err)
}
```

### 2. 记录日志

**Info 日志：**

```go
import (
    "context"
    "github.com/jessewkun/gocommon/logger"
)

ctx := context.Background()
logger.Info(ctx, "API", "用户登录成功, user_id: %d", userID)
```

**Error 日志：**

```go
err := someOperation()
if err != nil {
    logger.Error(ctx, "API", err)
}
```

**Debug 日志：**

```go
logger.Debug(ctx, "DEBUG", "调试信息: %+v", data)
```

**Warn 日志：**

```go
logger.Warn(ctx, "API", "警告信息: %s", message)
```

**Panic 和 Fatal 日志：**

```go
// Panic 会触发 panic
logger.Panic(ctx, "CRITICAL", "严重错误")

// Fatal 会调用 os.Exit(1)
logger.Fatal(ctx, "FATAL", "致命错误")
```

## 高级功能

### 结构化日志

支持使用字段记录结构化日志：

```go
// Info 日志带字段
logger.InfoWithField(ctx, "API", "用户操作", map[string]interface{}{
    "user_id": 123,
    "action":  "login",
    "ip":      "192.168.1.1",
})

// Error 日志带字段
logger.ErrorWithField(ctx, "API", "操作失败", map[string]interface{}{
    "user_id": 123,
    "error":   err.Error(),
})

// Debug 日志带字段
logger.DebugWithField(ctx, "DEBUG", "调试信息", map[string]interface{}{
    "request_id": "req-123",
    "duration":   100,
})

// Warn 日志带字段
logger.WarnWithField(ctx, "API", "警告信息", map[string]interface{}{
    "user_id": 123,
    "reason":   "rate_limit",
})
```

### 上下文传递

日志模块会自动从 context 中提取并记录以下信息：

- `trace_id`：链路追踪 ID（自动提取）
- `student_id`：学生 ID（自动提取）
- `teacher_id`：教师 ID（自动提取）
- `transparent_parameter`：配置的透传参数（自动提取）

**使用示例：**

```go
import (
    "context"
    "github.com/jessewkun/gocommon/constant"
)

// 在 context 中设置 trace_id
ctx := context.WithValue(context.Background(), constant.CtxTraceID, "trace-123")

// 记录日志，trace_id 会自动包含在日志中
logger.Info(ctx, "API", "处理请求")
```

### 透传参数

通过配置 `transparent_parameter`，可以从 context 中自动提取并记录自定义参数：

```toml
[log]
transparent_parameter = ["trace_id", "user_id", "request_id"]
```

```go
// 在 context 中设置参数
ctx := context.WithValue(ctx, "user_id", 123)
ctx := context.WithValue(ctx, "request_id", "req-456")

// 记录日志，user_id 和 request_id 会自动包含在日志中
logger.Info(ctx, "API", "处理请求")
```

### 报警集成

日志模块支持报警集成，当日志级别达到配置的报警级别时，会自动发送报警。

**配置报警级别：**

```toml
[log]
alarm_level = "warn"  # warn 或 error
```

**报警级别说明：**

- `warn`：Warn、Error、Fatal、Panic 级别日志会触发报警
- `error`：Error、Fatal、Panic 级别日志会触发报警

**注册报警器：**

```go
import (
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/alarm"
)

// 注册报警器（通常在应用启动时）
alerter := &alarm.Sender{}
logger.RegisterAlerter(alerter)
```

**强制发送报警：**

```go
// 无论报警级别如何，都发送报警
logger.InfoWithAlarm(ctx, "API", "重要信息需要报警")
```

**报警内容包含：**

- 日期时间
- 服务器 IP
- 主机名
- 日志标签（Tag）
- 日志消息
- 错误信息（如果有）
- 调用栈信息
- 请求路径（如果有）
- 学生 ID / 教师 ID（如果有）
- Trace ID（如果有）

### 日志轮转

日志模块基于 lumberjack 实现日志轮转：

- **按大小轮转**：当日志文件达到 `max_size`（MB）时自动轮转
- **按时间清理**：自动删除超过 `max_age` 天的日志文件
- **备份管理**：最多保留 `max_backup` 个备份文件

**日志文件命名：**

- 当前日志：`app.log`
- 轮转日志：`app.log.2025-01-01`, `app.log.2025-01-02`, ...

## 日志级别

### 日志级别说明

- **Debug**：调试信息，开发环境使用
- **Info**：一般信息，记录正常操作
- **Warn**：警告信息，需要注意但不影响运行
- **Error**：错误信息，操作失败但可以继续运行
- **Panic**：紧急错误，会触发 panic
- **Fatal**：致命错误，会调用 os.Exit(1)

### 日志级别配置

```toml
[log]
log_level = "info"  # debug, info, warn, error, fatal, panic
```

**级别过滤：**

- `debug`：记录所有级别
- `info`：记录 Info、Warn、Error、Fatal、Panic
- `warn`：记录 Warn、Error、Fatal、Panic
- `error`：记录 Error、Fatal、Panic
- `fatal`：记录 Fatal、Panic
- `panic`：只记录 Panic

## 完整示例

### 示例 1：基本日志记录

```go
package main

import (
    "context"
    "errors"
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/constant"
)

func main() {
    // 初始化日志（通常在配置加载后自动完成）
    if err := logger.Init(); err != nil {
        panic(err)
    }

    // 创建带 trace_id 的 context
    ctx := context.WithValue(context.Background(), constant.CtxTraceID, "trace-123")

    // 记录不同级别的日志
    logger.Debug(ctx, "MAIN", "应用启动")
    logger.Info(ctx, "MAIN", "应用运行中")
    logger.Warn(ctx, "MAIN", "警告信息")

    // 记录错误
    err := errors.New("操作失败")
    logger.Error(ctx, "MAIN", err)
}
```

### 示例 2：结构化日志

```go
package main

import (
    "context"
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/constant"
)

func handleRequest(ctx context.Context, userID int, action string) {
    // 使用结构化日志
    logger.InfoWithField(ctx, "API", "用户操作", map[string]interface{}{
        "user_id": userID,
        "action":  action,
        "ip":      getClientIP(ctx),
    })

    // 处理业务逻辑
    result, err := processAction(action)
    if err != nil {
        logger.ErrorWithField(ctx, "API", "操作失败", map[string]interface{}{
            "user_id": userID,
            "action":  action,
            "error":   err.Error(),
        })
        return
    }

    logger.InfoWithField(ctx, "API", "操作成功", map[string]interface{}{
        "user_id": userID,
        "action":  action,
        "result":  result,
    })
}
```

### 示例 3：报警集成

```go
package main

import (
    "context"
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/alarm"
)

func init() {
    // 注册报警器
    alerter := &alarm.Sender{}
    logger.RegisterAlerter(alerter)
}

func handleCriticalError(ctx context.Context, err error) {
    // 记录错误日志（如果配置了 alarm_level，会自动发送报警）
    logger.Error(ctx, "CRITICAL", err)

    // 或者强制发送报警
    logger.InfoWithAlarm(ctx, "CRITICAL", "重要信息需要立即通知")
}
```

### 示例 4：在 HTTP 处理器中使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 使用 Trace 中间件自动设置 trace_id
    r.Use(middleware.Trace())

    r.GET("/api/user", func(c *gin.Context) {
        // 使用 gin.Context 作为 context
        logger.Info(c, "API", "获取用户信息")

        // 处理业务逻辑
        user, err := getUser()
        if err != nil {
            logger.Error(c, "API", err)
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        logger.Info(c, "API", "用户信息获取成功, user_id: %d", user.ID)
        c.JSON(200, user)
    })

    return r
}
```

## 日志格式

日志以 JSON 格式输出，包含以下字段：

```json
{
  "level": "INFO",
  "datetime": "2025-01-01 12:00:00",
  "tag": "API",
  "msg": "用户登录成功",
  "host": "server-01",
  "ip": "192.168.1.100",
  "trace_id": "trace-123",
  "student_id": 123,
  "caller_line": "/path/to/file.go:123"
}
```

**字段说明：**

- `level`：日志级别
- `datetime`：时间戳
- `tag`：日志标签
- `msg`：日志消息
- `host`：主机名
- `ip`：服务器 IP
- `trace_id`：链路追踪 ID（如果有）
- `student_id` / `teacher_id`：用户 ID（如果有）
- `caller_line`：调用位置
- 其他自定义字段

## 注意事项

### 配置注意事项

1. **日志路径**：确保日志目录有写权限
2. **日志大小**：根据磁盘空间和日志量调整 `max_size`
3. **日志保留**：根据需求调整 `max_age` 和 `max_backup`
4. **日志级别**：生产环境建议使用 `info` 或 `warn`
5. **关闭日志**：`closed = true` 会关闭所有日志输出（用于测试）

### 性能注意事项

1. **零分配**：zap 库提供零分配日志记录，性能优异
2. **异步写入**：日志写入是异步的，不会阻塞业务逻辑
3. **字段类型**：使用合适的字段类型，避免不必要的序列化

### 报警注意事项

1. **报警级别**：合理配置报警级别，避免报警过多
2. **报警器注册**：确保在记录日志前注册报警器
3. **报警内容**：报警内容包含完整的上下文信息，便于排查问题

### 上下文传递注意事项

1. **Trace ID**：使用 `middleware.Trace()` 自动设置 trace_id
2. **异步任务**：在异步任务中使用 `common.CopyCtx()` 复制 context
3. **透传参数**：合理配置透传参数，避免记录过多信息

## 测试

运行测试：

```sh
go test ./logger -v
```

## 最佳实践

1. **日志级别使用**：
   - Debug：开发调试时使用
   - Info：记录正常业务流程
   - Warn：记录需要注意的情况
   - Error：记录错误，但不影响主流程
   - Panic/Fatal：记录严重错误，谨慎使用

2. **日志标签（Tag）**：
   - 使用有意义的标签，便于日志过滤和查询
   - 建议使用模块名或功能名作为标签

3. **结构化日志**：
   - 优先使用结构化日志，便于日志分析
   - 使用有意义的字段名

4. **上下文传递**：
   - 始终传递 context，确保 trace_id 等上下文信息被记录
   - 在异步任务中使用 `CopyCtx` 复制 context

5. **报警配置**：
   - 合理配置报警级别，避免报警过多
   - 确保报警器正确注册

6. **日志轮转**：
   - 根据日志量调整轮转参数
   - 定期清理旧日志文件

7. **性能优化**：
   - 生产环境使用合适的日志级别，减少日志量
   - 避免在循环中记录大量日志

## 获取原始 zap.Logger

如果需要直接使用 zap.Logger，可以获取：

```go
import "github.com/jessewkun/gocommon/logger"

// 获取 zap.Logger 实例
zapLogger := logger.Zap()
zapLogger.Info("直接使用 zap")
```

**注意事项：**

- 直接使用 zap.Logger 不会自动包含上下文信息
- 建议使用 logger 模块提供的函数，确保日志格式统一
