# Elasticsearch 模块

## 功能简介

本模块基于 [elastic/go-elasticsearch v8](https://github.com/elastic/go-elasticsearch) 深度封装，提供了健壮、可观测的 Elasticsearch 客户端管理功能。

-   ✅ 支持多实例连接管理
-   ✅ 启动时连接检查与自动重连
-   ✅ 统一的健康检查端点
-   ✅ 内置请求日志与慢查询监控
-   ✅ 优雅的连接关闭与资源释放
-   ✅ 灵活的 API，支持 `struct`, `[]byte`, `string`, `io.Reader` 等多种输入

## 依赖

-   Go 1.20+ (为了使用 `errors.Join`)
-   [github.com/elastic/go-elasticsearch/v8](https://github.com/elastic/go-elasticsearch)

## 配置示例

通过 `config` 包进行初始化，配置会自动加载到 `elasticsearch.Cfgs` 中。

**`config.toml` 中的配置:**
```toml
[elasticsearch.default]
addresses = ["http://127.0.0.1:9200"]
username = ""
password = ""
is_log = true          # 开启请求日志
slow_threshold = 200   # 慢查询阈值（毫秒）

[elasticsearch.another_cluster]
addresses = ["http://10.0.0.1:9200", "http://10.0.0.2:9200"]
is_log = true
slow_threshold = 500
```

**对应的 `type.go` 中 `Config` 结构体:**
```go
type Config struct {
	Addresses     []string `mapstructure:"addresses"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
	IsLog         bool     `mapstructure:"is_log"`         // 是否记录日志
	SlowThreshold int      `mapstructure:"slow_threshold"` // 慢查询阈值，单位毫秒
}
```

## 常用用法

### 1. 初始化
应用启动时，`config` 模块会自动调用 `elasticsearch.Init()`。

### 2. 获取客户端
```go
import "github.com/jessewkun/gocommon/db/elasticsearch"

// 获取名为 "default" 的 ES 客户端
client, err := elasticsearch.GetConn("default")
if err != nil {
    log.Fatalf("获取客户端失败: %v", err)
}
```

### 3. API 操作

新的 API 返回 `*esapi.Response`，你需要**手动关闭其 Body**。推荐使用 `defer res.Body.Close()`。

```go
ctx := context.Background()

// --- 索引管理 ---
const indexName = "my_test_index"

// 创建索引 (使用 map 定义 mapping)
mapping := map[string]interface{}{
    "mappings": {"properties": {"title": {"type": "text"}}},
}
res, err := client.CreateIndex(ctx, indexName, mapping)
if err != nil {
    // 错误处理...
} else {
    defer res.Body.Close()
    if res.IsError() {
        // API 返回错误...
    }
}

// --- 文档操作 ---
type MyDoc struct {
    Title string `json:"title"`
}

// 写入/更新文档 (使用 struct)
doc := MyDoc{Title: "Hello Elasticsearch!"}
res, err = client.Index(ctx, indexName, "doc_id_1", doc)
if err != nil {
    // ...
}
defer res.Body.Close()


// 查询文档
getRes, err := client.Get(ctx, indexName, "doc_id_1")
if err != nil {
    // ...
}
defer getRes.Body.Close()

if !getRes.IsError() {
    var docWrapper struct {
        Source MyDoc `json:"_source"`
    }
    // 使用辅助函数解码
    if err := elasticsearch.DecodeResponse(getRes.Body, &docWrapper); err != nil {
        // 解码失败...
    }
    fmt.Printf("查询成功: %+v\n", docWrapper.Source)
}

// 搜索 (使用 string)
query := `{"query":{"match_all":{}}}`
searchRes, err := client.Search(ctx, indexName, query)
if err != nil {
    // ...
}
defer searchRes.Body.Close()
// 读取响应...


// 删除文档
res, err = client.Delete(ctx, indexName, "doc_id_1")
// ...
defer res.Body.Close()

// 删除索引
res, err = client.DeleteIndex(ctx, indexName)
// ...
defer res.Body.Close()
```

### 4. 健康检查
```go
// 返回所有已配置实例的健康状态
healthStatusMap := elasticsearch.HealthCheck()
for name, status := range healthStatusMap {
    fmt.Printf("Cluster '%s': status=%s, latency=%dms\n", name, status.Status, status.Latency)
}
```

### 5. 关闭连接
应用退出前，可以调用 `Close` 来清理。
```go
elasticsearch.Close()
```

## 连接管理
本模块采用 `Manager` 模式管理所有连接实例。
-   **初始化**: `Init()` 函数会根据配置创建所有 ES 客户端，并进行连通性检查。任何失败的连接都会被记录并汇总返回。
-   **获取**: `GetConn(name)` 通过名称安全地获取一个已初始化的客户端。
-   **关闭**: `Close()` 逻辑上清空所有连接，以便垃圾回收。

## 日志和可观测性
当配置中 `is_log = true` 时，模块会自动注入一个日志中间件 (`loggingTransport`)。
-   **请求日志**: 每一次 ES 请求（包括方法、URL、耗时、状态码）都会被记录。
-   **慢查询**: 耗时超过 `slow_threshold`（毫秒）的请求会被标记为 `ES_SLOW_QUERY` 并以 `WARN` 级别记录。
-   **错误日志**: 请求失败或 ES 返回错误状态码时，会以 `ERROR` 级别记录，并包含部分请求和响应体以便调试。

## 测试
确保本地或目标环境已启动 Elasticsearch 服务，且测试代码中的地址正确。

运行测试：
```sh
go test ./db/elasticsearch -v
```
测试代码已被重构，分为 `Manager` 测试和 `DocumentFlow` 测试，更易于维护。
