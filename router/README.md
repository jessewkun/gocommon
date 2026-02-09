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

## WebSocket

提供统一的 WebSocket 升级与消息循环，支持可配置的跨域（CheckOrigin）和读超时。

### 接口

实现 `WebSocketHandler` 即可处理连接、消息与断开：

```go
type WebSocketHandler interface {
	OnConnect(c *gin.Context, conn *websocket.Conn) error
	OnMessage(c *gin.Context, messageType int, message []byte)
	OnDisconnect(c *gin.Context, conn *websocket.Conn)
	OnPing(c *gin.Context, conn *websocket.Conn)
}
```

### 默认用法

不配置跨域、使用默认 90 秒读超时：

```go
router.WsHandler(c, myHandler)
```

### 带配置用法（跨域、超时）

需要从业务配置传入允许的 Origin 或自定义校验时：

```go
cfg := &router.WsConfig{
	CheckOrigin:  router.CheckOriginFromAllowList(["http://test.com"]),
	ReadTimeout: 90 * time.Second, // 可选，不设则默认 90s
}
router.WsHandlerWithConfig(c, myHandler, cfg)
```

- `CheckOrigin` 为 nil 时使用 gorilla/websocket 默认：仅当请求无 Origin 或 Origin 与 Host 一致时通过。
- `CheckOriginFromAllowList(origins []string)` 可根据白名单生成校验函数；无 Origin 头的请求会放行。
