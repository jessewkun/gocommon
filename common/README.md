# 通用基础模块

## 功能简介

通用基础模块提供项目中常用的工具函数和类型定义，包括上下文管理、错误处理和环境判断等功能。

- ✅ **可扩展的上下文复制工具**：避免 goroutine 中 context 被取消，支持动态注册需传播的键。
- ✅ **自定义错误类型**：支持错误码和错误信息，明确区分系统错误和业务错误。
- ✅ **环境判断函数**：支持 debug、release、test 模式判断。
- ✅ **完善的错误处理机制**：支持错误包装和通过辅助函数判断错误码。

## 依赖

- `github.com/jessewkun/gocommon/config`：配置管理模块（用于环境判断）
- `github.com/jessewkun/gocommon/constant`：常量定义模块

## 上下文管理

### CopyCtx 函数

`CopyCtx` 用于复制一个新的 context，避免在 gin 框架中 HTTP 请求结束后，context 被 cancel，导致在请求中新开的 goroutine 中使用 context 时出现 `ctx canceled` 错误。

**函数签名：**

```go
func CopyCtx(ctx context.Context) context.Context
```

**功能特性：**

-   创建新的 `context.Background()`，不会被取消。
-   **动态复制关键上下文值**：通过注册机制，自动复制 `propagatedContextKeys` 中定义的键。
-   保留链路追踪信息，便于日志记录和问题排查。

**使用场景：**

在 HTTP 请求处理中启动异步任务（goroutine）时，需要使用 `CopyCtx` 复制 context，避免请求结束后 context 被取消。

**使用示例：**

```go
import (
    "context"
    "github.com/jessewkun/gocommon/common"
    "github.com/jessewkun/gocommon/constant"
)

func init() {
    // 可以在 init 函数中注册需要在 context 中传播的键
    common.RegisterPropagatedContextKey(constant.CtxUserID)
    common.RegisterPropagatedContextKey(constant.CtxTenantID) // 示例：注册一个新的键
}

func handleRequest(ctx context.Context) {
    // 在请求处理中启动异步任务
    go func() {
        // 使用 CopyCtx 复制 context，避免请求结束后被取消
        newCtx := common.CopyCtx(ctx)

        // 在异步任务中使用 newCtx
        doAsyncWork(newCtx)
    }()

    // 请求处理逻辑...
}

func doAsyncWork(ctx context.Context) {
    // 这个函数可能在 HTTP 请求结束后才执行
    // 使用 CopyCtx 复制的 context 不会被取消
    // logger.Info(ctx, "ASYNC", "执行异步任务",
    //     "tenant_id", ctx.Value(constant.CtxTenantID))
}
```

**注意事项：**

-   复制的 context 不会被取消，适合长时间运行的异步任务。
-   只复制通过 `RegisterPropagatedContextKey` 注册的键，其他值不会被复制。
-   注册键应在程序启动阶段完成（例如 `init()` 函数中）。

### RegisterPropagatedContextKey 函数

`RegisterPropagatedContextKey` 用于动态注册需要 `CopyCtx` 复制的上下文键。

**函数签名：**

```go
func RegisterPropagatedContextKey(key constant.ContextKey)
```

**功能特性：**

-   允许在运行时扩展 `CopyCtx` 传播的上下文键集合。
-   线程安全，可并发调用。
-   自动避免重复注册相同的键。

**使用示例（同上 `CopyCtx` 示例中的 `init()` 函数）：**

```go
import (
    "github.com/jessewkun/gocommon/common"
    "github.com/jessewkun/gocommon/constant"
)

func init() {
    common.RegisterPropagatedContextKey(constant.CtxUserID)
    common.RegisterPropagatedContextKey(constant.CtxTenantID)
}
```

## 错误处理

### CustomError 自定义错误类型

`CustomError` 提供了带错误码的自定义错误类型，支持错误包装。

**类型定义：**

```go
type CustomError struct {
    Code int   // 错误码
    Err  error // 原始错误
}
```

**功能特性：**

-   实现 `error` 接口。
-   实现 `errors.Unwrap` 接口，支持错误包装。
-   **明确区分系统错误和业务错误码区间**。

**使用示例：**

```go
import (
    "errors"
    "fmt"
    "github.com/jessewkun/gocommon/common"
)

// 创建自定义业务错误
err := errors.New("用户不存在")
customErr := common.NewBusinessError(10001, err)

// 使用错误
if customErr != nil {
    fmt.Printf("错误码: %d, 错误信息: %s\n", customErr.Code, customErr.Error())
}

// 错误包装
wrappedErr := common.NewBusinessError(10002, customErr)
if errors.Is(wrappedErr, customErr) { // errors.Is 仍然用于检查错误链中的实例
    fmt.Println("可以找到包装的错误")
}
```

### NewBusinessError 函数

创建业务自定义错误，业务错误码必须大于等于 10001。

**函数签名：**

```go
func NewBusinessError(code int, err error) CustomError
```

**错误码规范：**

-   **系统错误码**：< 10001（由系统使用，通过 `NewSystemError` 创建）
-   **业务错误码**：≥ 10001（由业务使用）

**使用示例：**

```go
// 正确的业务错误码
err := common.NewBusinessError(10001, errors.New("用户不存在"))
err2 := common.NewBusinessError(10002, errors.New("密码错误"))

// 错误码小于 10001 会 panic
// err := common.NewBusinessError(10000, errors.New("错误")) // 会 panic
```

### NewSystemError 函数

创建系统自定义错误，系统错误码必须小于 10001。

**函数签名：**

```go
func NewSystemError(code int, err error) CustomError
```

**使用示例：**

```go
// 正确的系统错误码
err := common.NewSystemError(1000, errors.New("数据库连接失败"))
err2 := common.NewSystemError(1001, errors.New("外部服务调用超时"))

// 错误码大于等于 10001 会 panic
// err := common.NewSystemError(10001, errors.New("错误")) // 会 panic
```

### 错误方法

**Error() 方法：**

返回错误信息，如果原始错误存在则返回原始错误信息，否则返回错误码。

```go
err := common.NewBusinessError(10001, errors.New("用户不存在"))
fmt.Println(err.Error()) // 输出: "用户不存在"

err2 := common.NewBusinessError(10002, nil)
fmt.Println(err2.Error()) // 输出: "code: 10002"
```

**String() 方法：**

返回更详细的错误信息，包括错误码和内部错误（如果存在）。

```go
err := common.NewBusinessError(10001, errors.New("用户不存在"))
fmt.Println(err.String()) // 输出: "code: 10001, err: 用户不存在"

err2 := common.NewBusinessError(10002, nil)
fmt.Println(err2.String()) // 输出: "code: 10002, err: nil"
```

**Unwrap() 方法：**

返回包装的原始错误，支持错误链解包。

```go
originalErr := errors.New("原始错误")
customErr := common.NewBusinessError(10001, originalErr)
unwrapped := customErr.Unwrap()
fmt.Println(unwrapped == originalErr) // true
```

### IsCode 辅助函数

`IsCode` 辅助函数用于判断错误链中是否存在指定错误码的 `CustomError`。

**函数签名：**

```go
func IsCode(err error, code int) bool
```

**使用场景：** 替代 `errors.Is(customErr, common.CustomError{Code: 10001})` 这种模糊的写法，提供清晰的错误码判断方式。

```go
import (
    "errors"
    "fmt"
    "github.com/jessewkun/gocommon/common"
)

func authenticate(username, password string) error {
    if username != "admin" {
        return common.NewBusinessError(10001, errors.New("用户不存在"))
    }
    if password != "password" {
        return common.NewBusinessError(10002, errors.New("密码错误"))
    }
    return nil
}

func main() {
    err := authenticate("guest", "123")
    if err != nil {
        if common.IsCode(err, 10001) {
            fmt.Println("用户不存在错误")
        } else if common.IsCode(err, 10002) {
            fmt.Println("密码错误")
        } else if common.IsCode(err, 1000) { // 示例：判断系统错误
            fmt.Println("错误：系统繁忙")
        } else {
            fmt.Printf("其他错误: %v\n", err)
        }
    } else {
        fmt.Println("登录成功！")
    }
}
```

## 环境判断

### 环境模式常量

```go
const (
    ModeDebug   = "debug"   // 开发环境
    ModeRelease = "release" // 生产环境
    ModeTest    = "test"    // 测试环境
)
```

### 环境判断函数

提供三个函数用于判断当前运行模式：

**IsDebug()：**

判断是否是 debug 模式（开发环境）。

```go
if common.IsDebug() {
    // 开发环境逻辑
    fmt.Println("当前是开发环境")
}
```

**IsRelease()：**

判断是否是 release 模式（生产环境）。

```go
if common.IsRelease() {
    // 生产环境逻辑
    fmt.Println("当前是生产环境")
}
```

**IsTest()：**

判断是否是 test 模式（测试环境）。

```go
if common.IsTest() {
    // 测试环境逻辑
    fmt.Println("当前是测试环境")
}
```

**使用示例：**

```go
import "github.com/jessewkun/gocommon/common"

func initLogger() {
    if common.IsDebug() {
        // 开发环境：输出详细日志
        // logger.SetLevel("debug")
    } else if common.IsRelease() {
        // 生产环境：只输出错误日志
        // logger.SetLevel("error")
    } else if common.IsTest() {
        // 测试环境：输出所有日志
        // logger.SetLevel("info")
    }
}

func connectDatabase() {
    var dsn string
    if common.IsDebug() {
        dsn = "localhost:3306"
    } else {
        dsn = "prod-db:3306"
    }
    // db.Connect(dsn)
}
```

**配置说明：**

环境模式通过 `config.Cfg.Mode` 配置，在配置文件中设置：

```toml
mode = "debug"  # 或 "release" 或 "test"
```

## 完整示例

### 示例 1：在异步任务中使用 CopyCtx 和动态键注册

```go
package main

import (
    "context"
    "github.com/jessewkun/gocommon/common"
    "github.com/jessewkun/gocommon/constant"
    "github.com/jessewkun/gocommon/logger"
    "time"
)

// 假设这是一个在某个包的 init 函数中注册的键
func init() {
    common.RegisterPropagatedContextKey(constant.CtxUserID)
    common.RegisterPropagatedContextKey(constant.CtxTenantID) // 示例：注册一个业务自定义键
}

func main() {
    ctx := context.Background()
    ctx = context.WithValue(ctx, constant.CtxTraceID, "trace-123")
    ctx = context.WithValue(ctx, constant.CtxUserID, uint64(456))
    ctx = context.WithValue(ctx, constant.CtxTenantID, "my-company")

    // 启动异步任务
    go func() {
        // 复制 context，避免请求结束后被取消
        newCtx := common.CopyCtx(ctx)

        // 执行异步任务
        processAsyncTask(newCtx)
    }()

    // 主逻辑继续执行...
    time.Sleep(100 * time.Millisecond) // 确保 goroutine 有时间执行
}

func processAsyncTask(ctx context.Context) {
    // 这个函数可能在 HTTP 请求结束后才执行
    logger.Info(ctx, "ASYNC", "处理异步任务",
        "trace_id", ctx.Value(constant.CtxTraceID),
        "user_id", ctx.Value(constant.CtxUserID),
        "tenant_id", ctx.Value(constant.CtxTenantID),
    )
}
```

### 示例 2：使用 CustomError 和 IsCode 处理业务错误

```go
package main

import (
    "errors"
    "fmt"
    "github.com/jessewkun/gocommon/common"
)

// 定义业务错误常量，使用 NewBusinessError
var (
    ErrUserNotFound    = common.NewBusinessError(10001, errors.New("用户不存在"))
    ErrInvalidPassword = common.NewBusinessError(10002, errors.New("密码错误"))
    ErrSystemBusy      = common.NewSystemError(1000, errors.New("系统繁忙"))
)

func login(username, password string) error {
    if username != "test_user" {
        return ErrUserNotFound
    }
    if password != "test_pass" {
        return ErrInvalidPassword
    }
    return nil
}

func handleLogin(username, password string) {
    err := login(username, password)
    if err != nil {
        if common.IsCode(err, 10001) {
            fmt.Println("错误：用户不存在")
        } else if common.IsCode(err, 10002) {
            fmt.Println("错误：密码错误")
        } else if common.IsCode(err, 1000) { // 示例：判断系统错误
            fmt.Println("错误：系统繁忙")
        } else {
            fmt.Printf("其他错误: %v\n", err)
        }
    } else {
        fmt.Println("登录成功！")
    }
}

func main() {
    handleLogin("test_user", "wrong_pass")
    handleLogin("unknown_user", "some_pass")
    handleLogin("test_user", "test_pass")
}
```

### 示例 3：根据环境执行不同逻辑

```go
package main

import (
    "context"
    "fmt"
    "github.com/jessewkun/gocommon/common"
    "github.com/jessewkun/gocommon/config" // 假设 config 包可用
)

func init() {
    // 为了示例，模拟配置加载
    config.Cfg.Mode = common.ModeDebug // 或者 common.ModeRelease, common.ModeTest
    fmt.Printf("当前环境模式：%s\n", config.Cfg.Mode)

    if common.IsDebug() {
        fmt.Println("开发环境初始化逻辑...")
        // logger.SetLevel("debug")
    } else if common.IsRelease() {
        fmt.Println("生产环境初始化逻辑...")
        // logger.SetLevel("error")
    } else if common.IsTest() {
        fmt.Println("测试环境初始化逻辑...")
        // logger.SetLevel("info")
    }
}

func connectDatabase() {
    var dsn string
    if common.IsDebug() {
        dsn = "localhost:3306"
    } else {
        dsn = "prod-db:3306"
    }
    fmt.Printf("连接数据库：%s\n", dsn)
    // db.Connect(dsn)
}

func main() {
    connectDatabase()
    // 实际业务逻辑...
}
```

## 注意事项

### CopyCtx 使用注意事项

1.  **动态复制**：`CopyCtx` 只复制通过 `RegisterPropagatedContextKey` 注册的上下文键。
2.  **注册时机**：应在程序启动阶段（如 `init()` 函数）完成所有键的注册。
3.  **不会被取消**：复制的 context 基于 `context.Background()`，不会被取消，适合长时间运行的异步任务。
4.  **链路追踪**：保留 `TraceID` 等关键信息，便于日志记录和问题排查。

### CustomError 使用注意事项

1.  **错误码规范**：
    *   业务错误码必须大于等于 10001，通过 `NewBusinessError` 创建。
    *   系统错误码必须小于 10001，通过 `NewSystemError` 创建。
    *   错误码的统一管理非常重要，建议建立错误码文档。
2.  **错误码唯一性**：建议为每个业务/系统错误定义唯一的错误码。
3.  **错误包装**：支持错误包装，可以构建错误链。
4.  **错误判断**：使用 `common.IsCode(err, code)` 判断错误类型和代码，避免直接比较 `CustomError` 实例或其 `Code` 字段。

<h3>环境判断注意事项</h3>

1.  **配置依赖**：环境判断函数依赖 `config.Cfg.Mode`，确保配置已正确加载。
2.  **模式值**：模式值必须是 "debug"、"release" 或 "test"（大小写敏感）。
3.  **默认值**：如果 `config.Cfg.Mode` 未配置，默认模式为 "debug"。
4.  **全局状态**：注意这些函数依赖全局配置状态，在并发测试中需谨慎处理以避免竞争条件。

<h2>测试</h2>

运行测试：

```sh
go test ./common -v
```

测试覆盖：

-   `CopyCtx` 上下文复制功能及动态键传播。
-   `RegisterPropagatedContextKey` 键注册与线程安全。
-   `CustomError` 错误类型（包括 `Error()`, `String()`, `Unwrap()`）。
-   `NewBusinessError` 和 `NewSystemError` 的创建及错误码范围检查。
-   `IsCode` 辅助函数。
-   环境判断函数。

<h2>与其他模块集成</h2>

<h3>与 logger 模块集成</h3>

`CopyCtx` 复制的 context 保留 `TraceID` 等信息，可以用于日志记录：

```go
newCtx := common.CopyCtx(ctx)
// logger.Info(newCtx, "MODULE", "日志信息")
```

<h3>与 response 模块集成</h3>

`CustomError` 和 `IsCode` 可以用于 API 响应错误处理：

```go
if err != nil {
    // if customErr, ok := err.(common.CustomError); ok { // 不再需要直接类型断言
    //     response.Custom(c, customErr.Code, customErr.Error(), nil) // 示例：使用 Custom 返回
    //     return
    // }
    // 使用 IsCode 进行判断和处理
    if common.IsCode(err, 10001) {
        response.Custom(c, 10001, "用户错误", nil)
        return
    }
    response.Error(c, err) // 默认处理
    return
}
```

<h3>与 config 模块集成</h3>

环境判断函数依赖 `config.Cfg.Mode`：

```go
// 在配置加载后使用
if common.IsDebug() {
    // 开发环境逻辑
}
```

<h2>最佳实践</h2>

1.  **异步任务使用 CopyCtx 并注册键**：
    -   在 HTTP 请求中启动 goroutine 时，始终使用 `CopyCtx` 复制 context。
    -   所有需要在异步任务中传播的 context 键都应通过 `RegisterPropagatedContextKey` 注册。
    -   保留 `TraceID` 等关键信息，便于日志追踪。
2.  **错误码管理**：
    -   为业务错误定义唯一的错误码（≥ 10001），通过 `NewBusinessError` 创建。
    -   为系统错误定义唯一的错误码（< 10001），通过 `NewSystemError` 创建。
    -   使用常量定义错误码，避免硬编码。
    -   建立错误码文档，便于维护。
3.  **错误处理**：
    -   使用 `CustomError` 统一业务错误格式。
    -   支持错误包装，构建错误链。
    -   使用 `common.IsCode(err, code)` 判断错误类型和代码。
4.  **环境判断**：
    -   根据环境执行不同的初始化逻辑。
    -   使用环境判断控制日志级别、数据库连接等。
