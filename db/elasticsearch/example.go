package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jessewkun/gocommon/logger"
)

// MyDoc is an example document structure.
type MyDoc struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func Example() {
	// The Example functions are automatically run by `go test`.
	// As this example requires a running Elasticsearch instance, we wrap the logic
	// in a separate function and call it from a test if needed.
	// This function body demonstrates the API usage.
	runExample()
}

func runExample() {
	// 初始化 logger
	cfg := logger.DefaultConfig()
	cfg.Path = "./test.log"
	logger.Cfg = cfg
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 设置 ES 配置
	Cfgs = map[string]*Config{
		"default": {
			Addresses:     []string{"http://localhost:9200"},
			Username:      "",
			Password:      "",
			IsLog:         true,
			SlowThreshold: 100, // 100ms
		},
	}

	// 初始化 ES 连接
	if err := Init(); err != nil {
		fmt.Println("ES连接失败:", err)
		return
	}
	defer Close()

	// 获取客户端
	client, err := GetConn("default")
	if err != nil {
		fmt.Println("获取客户端失败:", err)
		return
	}

	ctx := context.Background()
	const indexName = "test_example_index"

	// 0. 在开始前清理，确保环境干净
	_ = client.DeleteIndex(ctx, indexName)

	// 1. 健康检查 (使用全局函数)
	hc := HealthCheck()
	fmt.Printf("健康检查: status=%s\n", hc["default"].Status)

	// 2. 创建索引
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title":   map[string]string{"type": "text"},
				"content": map[string]string{"type": "text"},
			},
		},
	}
	if err := client.CreateIndex(ctx, indexName, mapping); err != nil {
		fmt.Println("创建索引失败:", err)
	} else {
		fmt.Println("索引创建成功")
	}

	// 3. 写入文档 (带上 refresh 参数，让文档立即可见)
	doc := MyDoc{Title: "Hello ES", Content: "Elasticsearch example content"}
	if err := client.Index(ctx, indexName, "1", doc, "true"); err != nil {
		fmt.Println("写入文档失败:", err)
	} else {
		fmt.Println("文档写入成功")
	}

	// 4. 查询文档
	var docWrapper struct {
		Source MyDoc `json:"_source"`
	}
	if found, err := client.Get(ctx, indexName, "1", &docWrapper); err != nil {
		fmt.Println("查询文档失败:", err)
	} else if !found {
		fmt.Println("未查询到文档")
	} else {
		fmt.Printf("查询到文档: %+v\n", docWrapper.Source)
	}

	// 5. 搜索
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]string{"title": "Hello"},
		},
	}
	var searchResult json.RawMessage // 使用 json.RawMessage 来捕获原始 JSON
	if err := client.Search(ctx, indexName, query, &searchResult); err != nil {
		fmt.Println("搜索失败:", err)
	} else {
		fmt.Println("搜索结果:", string(searchResult))
	}

	// 6. 删除文档
	if err := client.Delete(ctx, indexName, "1", ""); err != nil {
		fmt.Println("删除文档失败:", err)
	} else {
		fmt.Println("文档删除成功")
	}

	// 7. 删除索引
	if err := client.DeleteIndex(ctx, indexName); err != nil {
		fmt.Println("删除索引失败:", err)
	} else {
		fmt.Println("索引删除成功")
	}
}
