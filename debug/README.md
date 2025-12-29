# Debug 模块

`debug` 模块提供了一个简单、高性能且支持动态配置的调试日志系统。它允许开发者在代码中植入调试信息，并通过配置文件按需开启或关闭特定模块的日志输出，对线上排查问题非常有帮助。

## 特性

-   ✅ **模块化控制**：通过配置文件中的 `module` 列表，精确控制需要输出调试日志的模块。
-   ✅ **动态热重载**：支持运行时安全地修改配置文件并生效，无需重启服务。
-   ✅ **高性能设计**：
    -   内部使用 `map` 结构，检查模块是否开启的时间复杂度为 O(1)。
    -   对于未开启调试的模块，调用 `debug.Log()` 的开销极小，几乎为零。
-   ✅ **双输出模式**：
    -   `console`：直接将调试信息打印到标准输出，适合开发环境。
    -   `log`：将调试信息集成到项目的 `logger` 模块，以 `DEBUG` 级别输出，适合生产环境。
-   ✅ **并发安全**：内部使用读写锁保护配置的读取和重载，可在高并发场景下安全使用。
-   ✅ **简洁的 API**：只提供 `Log()` 和 `IsDebug()` 两个核心函数，易于理解和使用。

## 配置说明

在您的 `config.toml` (或其他格式的配置文件) 中添加 `[debug]` 部分。

```toml
# config.toml

[debug]
# 需要开启调试日志的模块列表。
# 这里的 "mysql", "http", "my_service" 都是示例，应替换为代码中实际使用的模块名。
module = ["mysql", "http", "my_service"]

# 输出模式: "console" 或 "log"。
# "console" - 打印到标准输出。
# "log" - 使用 gocommon/logger 组件输出。
mode = "console"
```

## 快速开始

### 输出调试日志

在代码中直接调用 `debug.Log()` 函数。

-   第一个参数是 `context`。
-   第二个参数是**模块名**（字符串），它将与配置文件中的 `module` 列表进行匹配。
-   后续参数与 `fmt.Printf` 一致。

```go
package main

import (
	"context"
	"time"
	"github.com/jessewkun/gocommon/debug"
	"github.com/jessewkun/gocommon/config" // 假设使用 config 包加载配置
)

func main() {
    // 1. 初始化配置 (通常在 main 函数开始时执行)
    // config.Init() 会自动加载配置文件并触发 debug 模块的配置加载
    if err := config.Init(); err != nil {
        // ... handle error
    }

    // 2. 在业务逻辑中使用 debug.Log
    ctx := context.Background()

    // 假设 "mysql" 在 debug.module 配置列表中
    // 这条日志将会被打印
    debug.Log(ctx, "mysql", "Slow query detected (%.2fms): %s", 123.45, "SELECT * FROM users WHERE id = 1")

    // 假设 "redis" 不在 debug.module 配置列表中
    // 这条日志将不会有任何输出，且调用开销极小
    debug.Log(ctx, "redis", "Cache key not found: %s", "user:123")

    // 模拟热重载：在真实应用中，这是通过修改配置文件自动触发的。
    // 如果现在将 "redis" 添加到配置文件的 module 列表中，
    // 在短暂的延迟后，下方的 debug.Log("redis", ...) 调用就会开始输出日志。
}
```

### 检查模块是否开启

在需要更复杂的条件判断时，可以使用 `debug.IsDebug()`。

```go
func processRequest(ctx context.Context) {
    if debug.IsDebug("http") {
        // 只有在 http 模块调试开启时，才执行相对耗时的调试准备工作
        headers := collectAllHeaders() // 假设这是一个耗时操作
        debug.Log(ctx, "http", "Request received with headers: %v", headers)
    }

    // ... 正常业务逻辑
}
```

## API 参考

-   `Log(ctx context.Context, module string, format string, v ...interface{})`

    如果 `module` 在配置中被启用，则按 `format` 格式化并输出调试日志。

-   `IsDebug(module string) bool`

    检查指定的 `module` 是否在配置中被启用。

## 设计与实现说明

为了保证高性能和安全性，`debug` 模块内部采用了以下设计：

-   **配置存储**: 启动和重载时，会将 `module` 列表转存到一个 `map[string]struct{}` 中。
-   **检查逻辑**: `IsDebug()` 和 `Log()` 中的检查操作是对 `map` 的一次查找，其时间复杂度为 O(1)。
-   **并发控制**: 使用 `sync.RWMutex` 对配置的读取和写入（热重载）进行加锁，确保在高并发环境下模块的正常工作。
-   **零开销**: 当一个模块的调试未开启时，`Log()` 函数在执行一次 `map` 查找后会立即返回，几乎没有 CPU 和内存开销。
