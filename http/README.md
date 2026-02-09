# HTTP 客户端模块

## 功能简介

HTTP 客户端模块提供简洁易用的 HTTP 请求封装，基于 `resty` 库实现，支持 GET、POST、流式 POST、文件上传、文件下载等常用操作。

- ✅ 简洁的 API 设计，易于使用
- ✅ 支持 GET、POST 请求
- ✅ 支持流式 POST（PostStream），按行回调，适用于 SSE 等场景
- ✅ 支持文件上传（字节流和文件路径两种方式）
- ✅ 支持文件下载
- ✅ 支持请求超时设置（全局和单次请求）
- ✅ 支持自动重试机制（可配置重试次数、等待时间）
- ✅ 支持 5xx 状态码重试（可选）
- ✅ 支持请求日志记录（可配置）
- ✅ 支持透传参数（从 context 自动提取并添加到请求头）
- ✅ 支持流式响应 buffer 配置（初始容量与单行最大长度）
- ✅ 支持配置热更新
- ✅ 提供 URL 查询参数构建工具
- ✅ 提供 Cookie 设置工具（支持跨域）

## 依赖

- `github.com/go-resty/resty/v2`：HTTP 客户端库
- `github.com/jessewkun/gocommon/config`：配置管理模块
- `github.com/jessewkun/gocommon/logger`：日志模块

## 配置说明

### Config 配置结构

```go
type Config struct {
    TransparentParameter []string `mapstructure:"transparent_parameter" json:"transparent_parameter"` // 透传参数列表
    IsLog                bool     `mapstructure:"is_log" json:"is_log"`                                 // 是否记录请求日志
}
```

### 配置示例

**`config.toml` 中的配置：**

```toml
[http]
transparent_parameter = ["X-User-ID", "X-Trace-ID", "X-Custom-ID"]
is_log = true
```

**`config.json` 中的配置：**

```json
{
  "http": {
    "transparent_parameter": ["X-User-ID", "X-Trace-ID", "X-Custom-ID"],
    "is_log": true
  }
}
```

### 配置项说明

- **transparent_parameter**：透传参数列表，配置后会自动从 context 中提取这些参数的值，并添加到请求头中。支持热更新，每次请求都会读取最新配置。
- **is_log**：是否记录请求日志，默认 `false`。开启后会记录请求 URL、响应数据、追踪信息、请求头等信息。

## 基本使用

### 1. 创建客户端

```go
import (
    "time"
    "github.com/jessewkun/gocommon/http"
)

// 创建基础客户端
client := http.NewClient(http.Option{
    Timeout: 10 * time.Second,
})

// 创建带重试的客户端
client := http.NewClient(http.Option{
    Timeout:            10 * time.Second,
    Retry:              3,
    RetryWaitTime:      1 * time.Second,
    RetryMaxWaitTime:   5 * time.Second,
    RetryWith5xxStatus: true, // 对 5xx 状态码进行重试
})

// 创建带默认请求头的客户端
client := http.NewClient(http.Option{
    Timeout: 10 * time.Second,
    Headers: map[string]string{
        "User-Agent": "MyApp/1.0",
        "Accept":     "application/json",
    },
})

// 禁用日志（覆盖配置）
isLog := false
client := http.NewClient(http.Option{
    Timeout: 10 * time.Second,
    IsLog:   &isLog,
})

// 流式请求 buffer 配置（PostStream 按行扫描时的初始容量与单行最大长度，为 0 时使用默认 64KB/1MB）
client := http.NewClient(http.Option{
    Timeout:                30 * time.Second,
    StreamBufferInitial:    128 * 1024,      // 128KB 初始
    StreamBufferMax:        2 * 1024 * 1024, // 2MB 单行最大
})
```

**Option 字段说明：**

| 字段 | 说明 |
|------|------|
| Headers | 默认请求头 |
| Timeout | 默认超时时间 |
| Retry | 最大重试次数 |
| RetryWaitTime / RetryMaxWaitTime | 重试等待时间 |
| RetryWith5xxStatus | 是否对 5xx 状态码重试 |
| IsLog | 是否记录请求日志，nil 表示不覆盖配置 |
| StreamBufferInitial | 流式响应行 buffer 初始容量（字节），默认 64KB |
| StreamBufferMax | 流式响应单行最大长度（字节），默认 1MB |

### 2. GET 请求

```go
import (
    "context"
    "github.com/jessewkun/gocommon/http"
)

ctx := context.Background()

// 基本 GET 请求
req := http.RequestGet{
    URL: "https://api.example.com/users",
}

resp, err := client.Get(ctx, req)
if err != nil {
    log.Printf("请求失败: %v", err)
    return
}

fmt.Printf("状态码: %d\n", resp.StatusCode)
fmt.Printf("响应体: %s\n", string(resp.Body))

// 带请求头的 GET 请求
req := http.RequestGet{
    URL: "https://api.example.com/users",
    Headers: map[string]string{
        "Authorization": "Bearer token123",
        "Accept":        "application/json",
    },
}

// 带超时的 GET 请求
req := http.RequestGet{
    URL:     "https://api.example.com/users",
    Timeout: 5 * time.Second, // 单次请求超时，覆盖客户端默认超时
}

resp, err := client.Get(ctx, req)
```

### 3. POST 请求

```go
// 发送 JSON 数据
user := map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
}

req := http.RequestPost{
    URL:     "https://api.example.com/users",
    Payload: user,
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
}

resp, err := client.Post(ctx, req)
if err != nil {
    log.Printf("请求失败: %v", err)
    return
}

fmt.Printf("状态码: %d\n", resp.StatusCode)
fmt.Printf("响应体: %s\n", string(resp.Body))
```

### 4. 流式 POST（PostStream）

适用于 SSE（Server-Sent Events）等流式响应，按行回调，不将整个响应体读入内存。

```go
err := client.PostStream(ctx, http.RequestPost{
    URL:     "https://api.example.com/stream",
    Payload: bodyBytes,
    Headers: map[string]string{
        "Content-Type": "application/json",
        "Accept":       "text/event-stream",
    },
    Timeout: 5 * time.Minute, // 流式请求建议较长超时
}, func(line []byte) error {
    // 每收到一行调用一次，line 为原始行内容（含 "data: " 前缀等）
    fmt.Println(string(line))
    return nil // 返回非 nil 会中止流式读取
})
if err != nil {
    log.Printf("流式请求失败: %v", err)
}
```

**说明：**

- 使用 `bufio.Scanner` 按行读取，行 buffer 大小由创建客户端时的 `StreamBufferInitial`、`StreamBufferMax` 决定（未配置时默认 64KB 初始、1MB 单行最大）。
- 单行超过 `StreamBufferMax` 会报错，可根据接口实际情况在 Option 中调大。

### 5. 文件上传

#### 方式一：使用字节流上传

```go
fileBytes := []byte("文件内容")
req := http.RequestUpload{
    URL:       "https://api.example.com/upload",
    FileBytes: fileBytes,
    Param:     "file",        // 文件参数名
    FileName:  "test.txt",    // 文件名
    Data: map[string]string{  // 额外的表单数据
        "description": "测试文件",
        "category":    "test",
    },
    Headers: map[string]string{
        "X-Custom-Header": "custom-value",
    },
}

resp, err := client.Upload(ctx, req)
```

#### 方式二：使用文件路径上传

```go
req := http.RequestUploadWithFilePath{
    URL:      "https://api.example.com/upload",
    FilePath: "/path/to/file.txt",
    FileName: "file.txt",
    Param:    "file",
    Data: map[string]string{
        "description": "文件描述",
    },
}

resp, err := client.UploadWithFilePath(ctx, req)
```

### 6. 文件下载

```go
req := http.RequestDownload{
    URL:      "https://api.example.com/download/file.pdf",
    FilePath: "/path/to/save/file.pdf",
    Headers: map[string]string{
        "Accept": "application/pdf",
    },
}

resp, err := client.Download(ctx, req)
if err != nil {
    log.Printf("下载失败: %v", err)
    return
}

fmt.Printf("下载成功，状态码: %d\n", resp.StatusCode)
// 文件已保存到 FilePath 指定的路径
```

## 高级功能

### 透传参数

透传参数功能可以自动从 context 中提取配置的参数值，并添加到请求头中。这对于在微服务架构中传递用户 ID、追踪 ID 等信息非常有用。

**配置透传参数：**

```toml
[http]
transparent_parameter = ["X-User-ID", "X-Trace-ID"]
```

**使用示例：**

```go
// 在 context 中设置参数
ctx := context.WithValue(context.Background(), "X-User-ID", "12345")
ctx = context.WithValue(ctx, "X-Trace-ID", "trace-67890")

// 发送请求时，这些参数会自动添加到请求头
req := http.RequestGet{
    URL: "https://api.example.com/users",
}

resp, err := client.Get(ctx, req)
// 请求头中会自动包含：
// X-User-ID: 12345
// X-Trace-ID: trace-67890
```

**特性：**

- 支持热更新：每次请求都会读取最新的配置，无需重启服务
- 自动提取：从 context 中自动提取参数值
- 自动添加：自动添加到请求头，无需手动设置

### 重试机制

客户端支持自动重试机制，可以在网络不稳定或服务器临时故障时提高请求成功率。

```go
client := http.NewClient(http.Option{
    Timeout:            10 * time.Second,
    Retry:              3,                    // 最大重试次数
    RetryWaitTime:      1 * time.Second,      // 重试等待时间
    RetryMaxWaitTime:   5 * time.Second,      // 最大重试等待时间
    RetryWith5xxStatus: true,                 // 是否对 5xx 状态码进行重试
})
```

**重试策略：**

- 默认情况下，只对网络错误进行重试
- 如果启用 `RetryWith5xxStatus`，会对 500-599 状态码进行重试
- 重试等待时间会逐渐增加，但不会超过 `RetryMaxWaitTime`

### 流式响应 Buffer

`PostStream` 按行扫描响应体时，使用客户端创建时配置的 buffer 大小：

- **StreamBufferInitial**：行 buffer 初始容量（字节），默认 64KB（`http.DefaultStreamBufferInitial`）。
- **StreamBufferMax**：单行允许的最大长度（字节），默认 1MB（`http.DefaultStreamBufferMax`）。

若接口返回的 SSE 行较长（例如单行 JSON 很大），可在创建客户端时增大上述两个值，避免 `buffer overflow`。

### 请求日志

开启日志记录后，每次请求都会记录详细信息：

```go
// 在配置文件中开启
[http]
is_log = true

// 或在创建客户端时覆盖
isLog := true
client := http.NewClient(http.Option{
    Timeout: 10 * time.Second,
    IsLog:   &isLog,
})
```

**日志内容：**

- 请求 URL
- 响应数据
- 追踪信息（请求耗时、DNS 解析时间等）
- 请求头

### 配置热更新

HTTP 模块支持配置热更新，无需重启服务即可应用新配置：

```go
// 配置变更后，会自动调用 Reload 方法
// 透传参数配置会立即生效，下次请求时使用新配置
```

**注意事项：**

- 所有配置项都被认为是安全的，可以进行热更新
- 热更新后，新创建的客户端会使用新配置
- 已创建的客户端会继续使用创建时的配置，但透传参数会读取最新配置

## 工具函数

### BuildQuery

构建 URL 查询参数字符串：

```go
params := map[string]interface{}{
    "name":  "张三",
    "age":   25,
    "email": "zhangsan@example.com",
}

query := http.BuildQuery(params)
// 输出: age=25&email=zhangsan%40example.com&name=%E5%BC%A0%E4%B8%89

url := "https://api.example.com/users?" + query
```

### SetCookie

设置 Cookie，支持跨域场景（保留 domain 前的点）：

```go
import (
    "net/http"
    "time"
    "github.com/jessewkun/gocommon/http"
)

http.SetCookie(
    response,                    // http.ResponseWriter
    "session_id",                // Cookie 名称
    "abc123",                    // Cookie 值
    24*time.Hour,                // 过期时间
    "/",                         // 路径
    ".example.com",              // 域名（支持以点开头）
    true,                        // Secure
    true,                        // HttpOnly
    http.SameSiteNoneMode,       // SameSite
)
```

**特性：**

- 保留 domain 前的点（Go 官方 `SetCookie` 会去掉）
- 支持跨域请求携带 Cookie
- 支持所有 Cookie 属性设置

## 响应结构

```go
type Response struct {
    StatusCode int             // HTTP 状态码
    Body       []byte          // 响应体（字节数组）
    Header     http.Header     // 响应头
    TraceInfo  resty.TraceInfo // 请求追踪信息
}
```

**TraceInfo 包含的信息：**

- `DNSLookup`：DNS 解析时间
- `ConnTime`：连接建立时间
- `TCPConnTime`：TCP 连接时间
- `TLSHandshake`：TLS 握手时间
- `ServerTime`：服务器处理时间
- `ResponseTime`：响应时间
- `TotalTime`：总耗时

**使用示例：**

```go
resp, err := client.Get(ctx, req)
if err != nil {
    return err
}

fmt.Printf("状态码: %d\n", resp.StatusCode)
fmt.Printf("总耗时: %v\n", resp.TraceInfo.TotalTime)
fmt.Printf("DNS 解析: %v\n", resp.TraceInfo.DNSLookup)
fmt.Printf("服务器处理: %v\n", resp.TraceInfo.ServerTime)

// 解析 JSON 响应
var data map[string]interface{}
json.Unmarshal(resp.Body, &data)
```

## 错误处理

HTTP 客户端会返回详细的错误信息：

```go
resp, err := client.Get(ctx, req)
if err != nil {
    // 错误可能是：
    // - 网络错误（连接失败、超时等）
    // - DNS 解析错误
    // - 请求被取消
    log.Printf("请求失败: %v", err)
    return
}

// 即使请求成功，也需要检查状态码
if resp.StatusCode != http.StatusOK {
    log.Printf("请求返回错误状态码: %d", resp.StatusCode)
    log.Printf("响应内容: %s", string(resp.Body))
}
```

## 注意事项

1. **超时设置**：
   - 客户端级别的超时：创建客户端时设置，适用于所有请求
   - 请求级别的超时：在请求结构中设置，会覆盖客户端默认超时
   - 如果都不设置，使用 resty 的默认超时

2. **重试机制**：
   - 默认情况下，只对网络错误进行重试
   - 启用 `RetryWith5xxStatus` 后，会对 5xx 状态码进行重试
   - 重试会增加请求总耗时，请根据业务需求合理配置

3. **日志记录**：
   - 开启日志会记录所有请求信息，包括请求头、响应体等
   - 注意保护敏感信息，避免在日志中泄露密码、Token 等

4. **透传参数**：
   - 参数名区分大小写
   - 如果 context 中没有对应的值，不会添加该请求头
   - 支持热更新，每次请求都会读取最新配置

5. **文件上传**：
   - 使用 `Upload` 方法时，文件内容会完全加载到内存
   - 对于大文件，建议使用 `UploadWithFilePath` 方法
   - 文件上传使用 multipart/form-data 格式

6. **文件下载**：
   - 文件会直接保存到指定路径
   - 如果文件已存在，会被覆盖
   - 确保有写入权限

7. **流式请求（PostStream）**：
   - 按行回调，不缓冲整个响应体，适合 SSE 等长连接
   - 行 buffer 由 Option 的 `StreamBufferInitial`、`StreamBufferMax` 控制，为 0 时使用默认 64KB/1MB
   - 单行超过 `StreamBufferMax` 会报错，需根据实际接口调大

8. **并发安全**：
   - 客户端实例是并发安全的，可以在多个 goroutine 中使用
   - 建议为每个服务或模块创建独立的客户端实例

## 测试

运行测试：

```sh
go test ./http -v
```

测试说明：

- 测试代码包含单元测试和集成测试
- 集成测试会请求真实的 API（如 httpbin.org、jsonplaceholder.typicode.com）
- 测试会验证各种场景：成功请求、超时、错误处理、重试、透传参数等

## 示例代码

完整的使用示例请参考本目录下的测试文件：

- `http_test.go`：包含各种场景的测试用例

## 与其他模块集成

### 与日志模块集成

HTTP 模块会自动使用日志模块记录请求信息：

```go
import (
    "github.com/jessewkun/gocommon/logger"
    "github.com/jessewkun/gocommon/http"
)

// 日志模块会自动记录 HTTP 请求
logger.Info(ctx, "发送 HTTP 请求")
```

### 与配置模块集成

HTTP 模块会自动注册到配置模块，支持配置热更新：

```go
import "github.com/jessewkun/gocommon/config"

// 配置加载时，HTTP 模块会自动初始化
cfg, err := config.Init("./config.toml")

// 配置热更新时，HTTP 模块会自动重新加载配置
```
