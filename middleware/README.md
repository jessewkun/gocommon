# 中间件模块

## 功能简介

中间件模块提供丰富的 Gin/HTTP 通用中间件，涵盖认证授权、限流控制、日志记录、异常处理、链路追踪、跨域处理、监控指标等功能。

- ✅ JWT 认证中间件，支持 token 刷新、黑名单管理
- ✅ 登录态检查中间件，支持自定义检查逻辑
- ✅ 全局限流和 IP 限流，支持智能清理
- ✅ 请求响应日志记录，支持敏感数据脱敏
- ✅ Panic 恢复机制，防止服务崩溃
- ✅ 链路追踪支持，自动生成和传递 trace_id
- ✅ CORS 跨域处理，支持动态源验证
- ✅ Prometheus 监控指标采集

## 依赖

- `github.com/gin-gonic/gin`：Gin Web 框架
- `github.com/jessewkun/gocommon/logger`：日志模块
- `github.com/jessewkun/gocommon/constant`：常量定义模块
- `github.com/jessewkun/gocommon/prometheus`：Prometheus 监控模块

## 认证授权

### JWT 认证中间件

`JwtAuth()` 提供完整的 JWT 认证中间件，支持 token 刷新、黑名单管理、自定义错误处理。

**函数签名：**

```go
func JwtAuth(secretKey string, options ...JwtAuthOption) gin.HandlerFunc
```

**功能特性：**

- JWT token 验证
- Token 刷新机制
- 黑名单管理（支持 token 撤销）
- 自定义错误处理
- 用户信息注入到 context

**使用示例：**

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 使用 JWT 认证中间件
    r.Use(middleware.JwtAuth("your-secret-key",
        middleware.WithTokenRefresh(true),
        middleware.WithBlacklistCheck(func(token string) bool {
            // 检查 token 是否在黑名单中
            return isTokenBlacklisted(token)
        }),
    ))

    // 需要认证的路由
    r.GET("/api/user", func(c *gin.Context) {
        // 从 context 中获取用户信息
        userID := c.GetString("user_id")
        c.JSON(200, gin.H{"user_id": userID})
    })

    return r
}
```

### 登录态检查中间件

`CheckLogin()` 和 `NeedLogin()` 提供灵活的登录态检查框架，支持自定义检查逻辑。

**函数签名：**

```go
func CheckLogin(checkFn func(*gin.Context) bool) gin.HandlerFunc
func NeedLogin(checkFn func(*gin.Context) bool) gin.HandlerFunc
```

**使用示例：**

```go
// 自定义登录检查函数
checkLogin := func(c *gin.Context) bool {
    // 从 session 或 cookie 中检查登录状态
    session := sessions.Default(c)
    userID := session.Get("user_id")
    return userID != nil
}

// 使用登录检查中间件
r.Use(middleware.CheckLogin(checkLogin))

// 或者使用 NeedLogin（失败时返回错误）
r.Use(middleware.NeedLogin(checkLogin))
```

### Token 管理

提供 token 创建和撤销功能：

```go
// 创建 JWT token
token, err := middleware.CreateJwtToken(userID, secretKey, expiration)

// 撤销 token（加入黑名单）
middleware.RevokeToken(token)
```

## 限流控制

### 全局限流

支持全局限流器，所有请求共享限流配额。

**函数签名：**

```go
func RateLimiter(rate float64, burst int) gin.HandlerFunc
```

**使用示例：**

```go
import "github.com/jessewkun/gocommon/middleware"

// 全局限流：每秒 100 个请求，突发 200 个
r.Use(middleware.RateLimiter(100, 200))
```

### IP 限流

为每个 IP 创建独立的限流器，支持 IP 白名单。

**函数签名：**

```go
func IPRateLimiter(rate float64, burst int, options ...IPRateLimiterOption) gin.HandlerFunc
```

**功能特性：**

- 每个 IP 独立限流
- IP 白名单支持
- 自动清理长时间未使用的限流器
- 可配置清理间隔

**使用示例：**

```go
// IP 限流：每个 IP 每秒 10 个请求，突发 20 个
r.Use(middleware.IPRateLimiter(10, 20,
    middleware.WithIPWhitelist([]string{"127.0.0.1", "::1"}),
    middleware.WithCleanupInterval(5 * time.Minute),
))

// 或者使用默认配置
r.Use(middleware.IPRateLimiter(10, 20))
```

**配置选项：**

- `WithIPWhitelist(ips []string)`：设置 IP 白名单
- `WithCleanupInterval(d time.Duration)`：设置清理间隔（默认 10 分钟）

## 日志记录

### IO 日志中间件

`IOLog()` 记录请求和响应的详细信息，支持敏感数据脱敏、大小限制和跳过响应体日志。

**函数签名：**

```go
func IOLog(config *IOLogConfig) gin.HandlerFunc
```

**功能特性：**

- 记录请求体、响应体、请求头、查询参数
- 敏感数据脱敏（支持正则表达式配置）
- 可配置请求体和响应体大小限制（默认 100KB）
- 请求体和响应体超过大小限制时记录提示信息
- 支持跳过特定路由的请求体或响应体日志记录（使用 `SkipBodyLog`）
- 记录客户端 IP、User-Agent 等信息
- 根据 HTTP 状态码智能选择日志级别

**IOLogConfig 配置结构：**

```go
type IOLogConfig struct {
    LogRequestBody      bool     // 是否记录请求体，默认 true
    LogResponseBody     bool     // 是否记录响应体，默认 true
    MaxRequestBodySize  int64    // 请求体大小限制（字节），默认 100KB
    MaxResponseBodySize int64    // 响应体大小限制（字节），默认 100KB
    SensitiveFields     []string // 需要脱敏的字段（正则表达式）
    LogHeaders          bool     // 是否记录请求头，默认 false
    LogQuery            bool     // 是否记录查询参数，默认 true
    LogPath             bool     // 是否记录路径，默认 true
    LogClientInfo       bool     // 是否记录客户端信息，默认 true
}
```

**使用示例：**

```go
import "github.com/jessewkun/gocommon/middleware"

// 使用默认配置（请求体和响应体大小限制为 100KB）
r.Use(middleware.IOLog(nil))

// 自定义配置
config := &middleware.IOLogConfig{
    LogRequestBody:      true,
    LogResponseBody:      true,
    MaxRequestBodySize:  200 * 1024,  // 200KB
    MaxResponseBodySize: 200 * 1024,  // 200KB
    SensitiveFields: []string{
        `(?i)password`,
        `(?i)token`,
        `(?i)secret`,
    },
    LogHeaders:    false,
    LogQuery:      true,
    LogPath:       true,
    LogClientInfo: true,
}
r.Use(middleware.IOLog(config))
```

**跳过请求体或响应体日志：**

对于请求体或响应体较大的接口（如文件下载、文件上传等），可以使用 `SkipBodyLog()` 跳过日志记录：

**SkipBodyType 类型说明：**

- `SkipAllBody`：跳过请求体和响应体日志
- `SkipRequestBody`：只跳过请求体日志
- `SkipResponseBody`：只跳过响应体日志

**使用示例：**

```go
// 跳过响应体日志（文件下载接口）
router.GET("/api/export", middleware.SkipBodyLog(middleware.SkipResponseBody), exportHandler)
router.GET("/api/download/:file", middleware.SkipBodyLog(middleware.SkipResponseBody), downloadHandler)

// 跳过请求体日志（文件上传接口）
router.POST("/api/upload", middleware.SkipBodyLog(middleware.SkipRequestBody), uploadHandler)
router.PUT("/api/batch", middleware.SkipBodyLog(middleware.SkipRequestBody), batchHandler)

// 同时跳过请求体和响应体日志（大文件上传接口）
router.POST("/api/upload-large", middleware.SkipBodyLog(middleware.SkipAllBody), uploadLargeHandler)

// 路由组跳过响应体日志
exportGroup := router.Group("/api/export", middleware.SkipBodyLog(middleware.SkipResponseBody))
exportGroup.GET("/data", exportHandler)
exportGroup.GET("/report", reportHandler)

// 或使用 Use 方法
exportGroup := router.Group("/api/export")
exportGroup.Use(middleware.SkipBodyLog(middleware.SkipResponseBody))
exportGroup.GET("/data", exportHandler)
```

**大小限制行为：**

- **请求体**：超过 `MaxRequestBodySize` 时，记录 `"[请求体超过大小限制]"`
- **响应体**：超过 `MaxResponseBodySize` 时，记录 `"[响应体超过大小限制]"`
- **响应大小**：`response_length` 字段始终记录，表示实际响应大小
- **限制为 0**：设置为 0 表示不限制大小

**日志级别：**

- **5xx 错误**：使用 `ErrorWithField` 记录
- **其他状态码**：使用 `InfoWithField` 记录

## 异常处理

### Panic 恢复中间件

`Recovery()` 自动捕获和恢复 panic，防止服务崩溃。

**函数签名：**

```go
func Recovery() gin.HandlerFunc
```

**功能特性：**

- 自动捕获 panic
- 记录详细的 panic 堆栈信息
- 调试模式下输出 panic 信息到控制台
- 返回系统错误响应，保证服务可用性

**使用示例：**

```go
import "github.com/jessewkun/gocommon/middleware"

// 使用恢复中间件（通常放在最前面）
r.Use(middleware.Recovery())
```

**注意事项：**

- 建议将 Recovery 中间件放在最前面，确保能捕获所有 panic
- 生产环境会自动记录错误日志，不会输出到控制台

## 链路追踪

### 链路追踪中间件

`Trace()` 自动生成或从请求头获取 trace_id，并传递到 context 中。

**函数签名：**

```go
func Trace() gin.HandlerFunc
```

**功能特性：**

- 自动生成 trace_id（如果请求头中没有）
- 从请求头获取 trace_id（如果存在）
- 将 trace_id 和请求路径传递到 context
- 在响应头中添加服务器主机名标识

**使用示例：**

```go
import "github.com/jessewkun/gocommon/middleware"

// 使用链路追踪中间件
r.Use(middleware.Trace())

// 在处理器中使用 trace_id
r.GET("/api/test", func(c *gin.Context) {
    traceID := c.GetString("trace_id")
    logger.Info(c, "API", "处理请求, trace_id: %s", traceID)
})
```

## 跨域处理

### CORS 中间件

`CORS()` 支持自定义允许的源、方法、请求头，支持动态源验证。

**函数签名：**

```go
func CORS(options ...CORSOption) gin.HandlerFunc
```

**功能特性：**

- 自定义允许的源、方法、请求头
- 动态源验证（支持函数验证）
- 支持跨域请求携带凭证
- 预检请求处理

**使用示例：**

```go
import "github.com/jessewkun/gocommon/middleware"

// 使用默认配置（允许所有源）
r.Use(middleware.CORS())

// 自定义配置
r.Use(middleware.CORS(
    middleware.WithAllowedOrigins([]string{"https://example.com", "https://app.example.com"}),
    middleware.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
    middleware.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
    middleware.WithAllowCredentials(true),
))

// 动态源验证
r.Use(middleware.CORS(
    middleware.WithOriginValidator(func(origin string) bool {
        // 动态验证源
        return isValidOrigin(origin)
    }),
))
```

**配置选项：**

- `WithAllowedOrigins(origins []string)`：设置允许的源列表
- `WithAllowedMethods(methods []string)`：设置允许的 HTTP 方法
- `WithAllowedHeaders(headers []string)`：设置允许的请求头
- `WithAllowCredentials(allow bool)`：是否允许携带凭证
- `WithOriginValidator(validator func(string) bool)`：设置动态源验证函数

## 监控指标

### Prometheus 中间件

`Prometheus()` 记录请求统计和性能指标，支持 Prometheus 格式导出。

**函数签名：**

```go
func Prometheus() gin.HandlerFunc
```

**功能特性：**

- 记录请求总数、响应状态码
- 记录请求处理时长分布
- 智能识别注册的路由，避免动态路径产生过多指标
- 支持 Prometheus 格式导出

**使用示例：**

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// 使用 Prometheus 中间件
r.Use(middleware.Prometheus())

// 暴露 Prometheus 指标端点
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

**指标说明：**

- `http_requests_total`：请求总数（按方法、路径、状态码分组）
- `http_request_duration_seconds`：请求处理时长分布（支持 P50/P95/P99 分位数）

## 完整示例

### 示例 1：完整的中间件配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 1. 异常恢复（放在最前面）
    r.Use(middleware.Recovery())

    // 2. 链路追踪
    r.Use(middleware.Trace())

    // 3. CORS 跨域处理
    r.Use(middleware.CORS(
        middleware.WithAllowedOrigins([]string{"https://example.com"}),
        middleware.WithAllowCredentials(true),
    ))

    // 4. 请求日志记录
    r.Use(middleware.IOLog(nil)) // 使用默认配置

    // 或者自定义配置
    // ioLogConfig := &middleware.IOLogConfig{
    //     MaxRequestBodySize:  200 * 1024, // 200KB
    //     MaxResponseBodySize: 200 * 1024, // 200KB
    //     SensitiveFields: []string{"password", "token"},
    // }
    // r.Use(middleware.IOLog(ioLogConfig))

    // 5. 限流控制
    r.Use(middleware.IPRateLimiter(100, 200))

    // 6. Prometheus 监控
    r.Use(middleware.Prometheus())

    // 公开路由
    public := r.Group("/api/public")
    {
        public.POST("/login", loginHandler)
    }

    // 需要认证的路由
    protected := r.Group("/api")
    protected.Use(middleware.JwtAuth("your-secret-key"))
    {
        protected.GET("/user", getUserHandler)
        protected.POST("/user", updateUserHandler)
    }

    return r
}
```

### 示例 2：自定义认证检查

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 自定义登录检查
    checkLogin := func(c *gin.Context) bool {
        // 从 session 检查
        session := sessions.Default(c)
        userID := session.Get("user_id")
        if userID == nil {
            return false
        }

        // 可以添加更多检查逻辑
        // 如检查用户状态、权限等

        return true
    }

    // 使用自定义登录检查
    r.Use(middleware.NeedLogin(checkLogin))

    // 路由处理...

    return r
}
```

### 示例 3：IP 白名单限流

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // IP 限流，内网 IP 不受限制
    r.Use(middleware.IPRateLimiter(10, 20,
        middleware.WithIPWhitelist([]string{
            "127.0.0.1",
            "::1",
            "192.168.0.0/16",
            "10.0.0.0/8",
        }),
        middleware.WithCleanupInterval(5 * time.Minute),
    ))

    return r
}
```

### 示例 4：跳过请求体和响应体日志

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jessewkun/gocommon/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.Default()

    // 使用 IOLog 中间件
    r.Use(middleware.IOLog(nil))

    // 普通接口，正常记录请求体和响应体
    r.GET("/api/user", getUserHandler)
    r.POST("/api/user", updateUserHandler)

    // 文件下载接口，跳过响应体日志（响应体太大）
    r.GET("/api/export", middleware.SkipBodyLog(middleware.SkipResponseBody), exportHandler)
    r.GET("/api/download/:file", middleware.SkipBodyLog(middleware.SkipResponseBody), downloadHandler)

    // 文件上传接口，跳过请求体日志（请求体太大）
    r.POST("/api/upload", middleware.SkipBodyLog(middleware.SkipRequestBody), uploadHandler)

    // 大文件上传接口，同时跳过请求体和响应体日志
    r.POST("/api/upload-large", middleware.SkipBodyLog(middleware.SkipAllBody), uploadLargeHandler)

    // 批量上传接口，只跳过请求体日志
    r.POST("/api/batch-upload", middleware.SkipBodyLog(middleware.SkipRequestBody), batchUploadHandler)

    // 使用路由组跳过响应体日志
    exportGroup := r.Group("/api/export", middleware.SkipBodyLog(middleware.SkipResponseBody))
    exportGroup.GET("/data", exportDataHandler)
    exportGroup.GET("/report", exportReportHandler)

    // 使用路由组跳过请求体日志
    uploadGroup := r.Group("/api/upload", middleware.SkipBodyLog(middleware.SkipRequestBody))
    uploadGroup.POST("/file", uploadFileHandler)
    uploadGroup.POST("/image", uploadImageHandler)

    // 使用路由组同时跳过请求体和响应体日志
    largeUploadGroup := r.Group("/api/upload-large")
    largeUploadGroup.Use(middleware.SkipBodyLog(middleware.SkipAllBody))
    largeUploadGroup.POST("/file", uploadLargeFileHandler)

    return r
}
```

## 中间件执行顺序

建议的中间件执行顺序：

1. **Recovery**：异常恢复（最前面）
2. **Trace**：链路追踪
3. **CORS**：跨域处理
4. **IOLog**：请求日志记录
5. **RateLimiter / IPRateLimiter**：限流控制
6. **Prometheus**：监控指标
7. **JwtAuth / CheckLogin**：认证授权（按需使用）

## 注意事项

### JWT 认证

1. **Secret Key 安全**：确保 secret key 的安全性，不要硬编码
2. **Token 过期时间**：合理设置 token 过期时间
3. **黑名单管理**：实现持久化的黑名单存储（如 Redis）
4. **Token 刷新**：建议实现 token 刷新机制

### 限流控制

1. **限流参数**：根据实际业务调整限流速率和突发值
2. **IP 白名单**：合理配置内网 IP 白名单
3. **清理间隔**：根据内存使用情况调整清理间隔
4. **监控告警**：监控限流触发情况，及时调整参数

### 日志记录

1. **敏感数据**：确保配置所有敏感字段进行脱敏
2. **日志大小**：合理设置请求体和响应体大小限制（默认 100KB）
   - 超过限制时，会记录提示信息而不是实际内容
   - 可通过 `SkipBodyLog()` 完全跳过请求体或响应体日志
3. **跳过请求体或响应体日志**：
   - 一般要求都记录请求体和响应体，但对于较大的接口可以跳过
   - 使用 `SkipBodyLog(SkipRequestBody)` 跳过请求体日志
   - 使用 `SkipBodyLog(SkipResponseBody)` 跳过响应体日志
   - 使用 `SkipBodyLog(SkipAllBody)` 同时跳过请求体和响应体日志
   - 跳过请求体日志时，请求体仍会被读取（因为需要在 `c.Next()` 之前读取），但不会记录到日志中
   - 跳过响应体日志时，响应体不会被获取和记录
   - 其他请求信息（路径、查询参数、请求头等）和响应大小（`response_length`）仍会记录
5. **性能影响**：日志记录会影响性能，生产环境建议只记录关键信息
6. **日志存储**：确保日志存储和查询性能

### 异常处理

1. **错误信息**：生产环境不要暴露详细的错误信息
2. **日志记录**：确保 panic 信息被正确记录
3. **监控告警**：监控 panic 发生频率，及时处理

### 链路追踪

1. **Trace ID 传递**：确保在异步任务中使用 `CopyCtx` 复制 context
2. **日志关联**：使用 trace_id 关联所有相关日志
3. **性能影响**：trace_id 生成和传递的性能影响很小

### CORS 配置

1. **源验证**：生产环境严格配置允许的源
2. **凭证支持**：需要携带凭证时设置 `AllowCredentials`
3. **预检请求**：理解 CORS 预检请求机制

### Prometheus 监控

1. **指标标签**：避免使用动态路径产生过多指标
2. **指标暴露**：确保 `/metrics` 端点安全访问
3. **指标查询**：使用 PromQL 查询和分析指标

## 测试

运行测试：

```sh
go test ./middleware -v
```

测试覆盖：

- 限流器功能测试
- 中间件集成测试

## 与其他模块集成

### 与 logger 模块集成

中间件自动使用 logger 模块记录日志：

```go
// IOLog 中间件自动记录请求响应日志
r.Use(middleware.IOLog())

// Recovery 中间件自动记录 panic 日志
r.Use(middleware.Recovery())
```

### 与 prometheus 模块集成

Prometheus 中间件自动使用 prometheus 模块采集指标：

```go
r.Use(middleware.Prometheus())
```

### 与 constant 模块集成

中间件使用 constant 模块定义的上下文键：

```go
// Trace 中间件使用 constant.CtxTraceID
traceID := c.GetString(string(constant.CtxTraceID))
```

## 最佳实践

1. **中间件顺序**：按照建议的顺序配置中间件
2. **错误处理**：使用 Recovery 中间件防止服务崩溃
3. **日志记录**：合理配置日志记录，避免记录过多信息
4. **限流配置**：根据实际业务调整限流参数
5. **安全配置**：生产环境严格配置 CORS 和认证
6. **监控告警**：使用 Prometheus 监控关键指标
7. **性能优化**：避免在中间件中执行耗时操作
