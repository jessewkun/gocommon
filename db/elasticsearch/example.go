package elasticsearch

import (
	"context"
	"fmt"
	"log"

	gocommonlog "github.com/jessewkun/gocommon/logger"
)

func Example() {
	// 初始化 logger
	cfg := gocommonlog.DefaultConfig()
	cfg.Path = "./test.log"
	cfg.MaxSize = 100
	cfg.MaxAge = 30
	cfg.MaxBackup = 10
	cfg.AlarmLevel = "warn"
	gocommonlog.Cfg = cfg
	if err := gocommonlog.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 设置 ES 配置
	Cfgs = map[string]*Config{
		"default": {
			Addresses: []string{"http://localhost:9200"},
			Username:  "",
			Password:  "",
		},
	}

	// 初始化 ES 连接
	if err := Init(); err != nil {
		fmt.Println("ES连接失败:", err)
		return
	}

	// 获取客户端
	client, err := GetConn("default")
	if err != nil {
		fmt.Println("获取客户端失败:", err)
		return
	}

	ctx := context.Background()

	// 1. 健康检查
	hc := client.HealthCheck()
	fmt.Println("健康检查:", hc)

	// 2. 创建索引
	mapping := `{"mappings":{"properties":{"title":{"type":"text"},"content":{"type":"text"}}}}`
	err = client.CreateIndex(ctx, "test_index", mapping)
	if err != nil {
		fmt.Println("创建索引失败:", err)
	} else {
		fmt.Println("索引创建成功")
	}

	// 3. 写入文档
	doc := `{"title":"Hello ES","content":"Elasticsearch example content"}`
	err = client.Index(ctx, "test_index", "1", doc)
	if err != nil {
		fmt.Println("写入文档失败:", err)
	} else {
		fmt.Println("文档写入成功")
	}

	// 4. 查询文档
	res, err := client.Get(ctx, "test_index", "1")
	if err != nil {
		fmt.Println("查询文档失败:", err)
	} else {
		fmt.Println("查询文档:", res)
	}

	// 5. 搜索
	query := `{"query":{"match":{"title":"Hello"}}}`
	searchRes, err := client.Search(ctx, "test_index", query)
	if err != nil {
		fmt.Println("搜索失败:", err)
	} else {
		fmt.Println("搜索结果:", searchRes)
	}

	// 6. 删除文档
	err = client.Delete(ctx, "test_index", "1")
	if err != nil {
		fmt.Println("删除文档失败:", err)
	} else {
		fmt.Println("文档删除成功")
	}

	// 7. 删除索引
	err = client.DeleteIndex(ctx, "test_index")
	if err != nil {
		fmt.Println("删除索引失败:", err)
	} else {
		fmt.Println("索引删除成功")
	}
}
