# router 模块

`router` 模块用于向 Gin 引擎注册一组通用的、标准化的系统级路由。

## 功能

`RegisterSystemRoutes` 函数会向 Gin 引擎中注册以下路由：

1.  **健康检查 (Health Check)**
    -   **Endpoint**: `GET /health/ping`
    -   **功能**: 返回一个简单的 `pong` 字符串，用于服务存活探测。

2.  **Prometheus 指标 (Metrics)**
    -   **Endpoint**: `GET /metrics`
    -   **功能**: 暴露应用的 Prometheus 指标，以便监控系统进行抓取。

3.  **性能分析 (pprof)**
    -   **Endpoint**: `GET /debug/pprof/*`
    -   **功能**: 提供 Go 语言标准的性能分析工具 pprof 的各个端点。
    -   **安全**: **这是一个受保护的路由组**。为了安全起见，它内置了 IP 白名单限流中间件，默认**仅允许本地 IP (`127.0.0.1`, `::1`) 访问**。任何来自其他 IP 的请求都将被拒绝。

## 快速上手

在你的 Gin 应用初始化阶段，调用 `RegisterSystemRoutes` 即可。

### 使用示例

```go
package main

import (
    "github.com/jessewkun/gocommon/router"
    "github.com/gin-gonic/gin"
)

func main() {
    // 1. 创建 Gin 引擎
    r := gin.Default()

    // 2. 注册所有系统路由
    router.RegisterSystemRoutes(r)

    // 3. 注册你自己的业务路由...
    // r.GET("/my-api", ...)

    // 4. 启动服务
    if err := r.Run(":8080"); err != nil {
        panic(err)
    }
}
```

现在，你的服务就已经拥有了 `/health/ping`、`/metrics` 和仅限本地访问的 `/debug/pprof/*` 路由。
