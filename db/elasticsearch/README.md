# Elasticsearch 模块

## 功能简介

本模块基于 [elastic/go-elasticsearch v8](https://github.com/elastic/go-elasticsearch) 封装，提供了 Elasticsearch 的常用操作，包括：

-   客户端初始化
-   健康检查
-   索引管理（创建、删除、判断是否存在、获取 mapping）
-   文档操作（写入/更新、查询、删除、搜索）

## 依赖

-   Go 1.18+
-   [github.com/elastic/go-elasticsearch/v8](https://github.com/elastic/go-elasticsearch)

## 配置示例

```go
cfg := Config{
    Addresses: []string{"http://127.0.0.1:9200"}, // ES 服务地址
    Username:  "", // 如有安全认证请填写
    Password:  "",
}
client, err := NewClient(cfg)
if err != nil {
    panic(err)
}
```

## 常用用法

```go
ctx := context.Background()

// 健康检查
health := client.HealthCheck()
fmt.Println(health)

// 创建索引
mapping := `{"mappings":{"properties":{"title":{"type":"text"}}}}`
err = client.CreateIndex(ctx, "test_index", mapping)

// 判断索引是否存在
exists, err := client.IndexExists(ctx, "test_index")

// 获取索引 mapping
m, err := client.GetIndexMapping(ctx, "test_index")

// 写入文档
err = client.Index(ctx, "test_index", "1", `{"title":"Hello"}`)

// 查询文档
res, err := client.Get(ctx, "test_index", "1")

// 搜索
query := `{"query":{"match":{"title":"Hello"}}}`
searchRes, err := client.Search(ctx, "test_index", query)

// 删除文档
err = client.Delete(ctx, "test_index", "1")

// 删除索引
err = client.DeleteIndex(ctx, "test_index")
```

## 测试

确保本地或目标环境已启动 Elasticsearch 服务，且 `Addresses` 配置正确。

运行测试：

```sh
go test ./db/elasticsearch -v
```

## 其他说明

-   支持多节点集群，直接在 `Addresses` 填写多个地址即可。
-   如需更多高级用法（如批量、聚合、DSL 构建等），可参考官方文档或扩展本模块。
