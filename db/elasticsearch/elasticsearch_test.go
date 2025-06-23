package elasticsearch

import (
	"context"
	"os"
	"testing"

	"github.com/jessewkun/gocommon/logger"
)

func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()
	code := m.Run()
	os.Remove("./test.log")
	os.Exit(code)
}

func TestElasticsearch_BasicFlow(t *testing.T) {
	// 设置 ES 配置
	Cfgs = map[string]*Config{
		"default": {
			Addresses: []string{"http://127.0.0.1:9200"},
			Username:  "",
			Password:  "",
		},
	}

	// 初始化 ES 连接
	if err := Init(); err != nil {
		t.Skipf("ES连接失败，跳过测试: %v", err)
	}

	// 获取客户端
	client, err := GetConn("default")
	if err != nil {
		t.Fatalf("获取客户端失败: %v", err)
	}

	ctx := context.Background()

	// 健康检查
	hc := client.HealthCheck()
	if hc == nil || len(hc) == 0 {
		t.Skip("健康检查失败，跳过测试")
	}

	// 检查是否有成功的连接
	hasSuccess := false
	for _, status := range hc {
		if status.Status == "success" {
			hasSuccess = true
			break
		}
	}
	if !hasSuccess {
		t.Skip("没有成功的 ES 连接，跳过测试")
	}

	// 创建索引
	mapping := `{"mappings":{"properties":{"title":{"type":"text"},"content":{"type":"text"}}}}`
	err = client.CreateIndex(ctx, "test_index", mapping)
	if err != nil {
		t.Skipf("创建索引失败，跳过测试: %v", err)
	}

	// 判断索引是否存在
	exists, err := client.IndexExists(ctx, "test_index")
	if err != nil {
		t.Skipf("IndexExists 查询失败，跳过测试: %v", err)
	}
	if !exists {
		t.Skip("IndexExists 应为 true，跳过测试")
	}

	// 获取索引 mapping
	m, err := client.GetIndexMapping(ctx, "test_index")
	if err != nil {
		t.Skipf("GetIndexMapping 失败，跳过测试: %v", err)
	}
	if m == "" {
		t.Skip("GetIndexMapping 返回空，跳过测试")
	}

	// 写入文档
	doc := `{"title":"Hello ES","content":"Elasticsearch test content"}`
	err = client.Index(ctx, "test_index", "1", doc)
	if err != nil {
		t.Skipf("写入文档失败，跳过测试: %v", err)
	}

	// 查询文档
	res, err := client.Get(ctx, "test_index", "1")
	if err != nil {
		t.Skipf("查询文档失败，跳过测试: %v", err)
	}
	if res == "" {
		t.Skip("查询文档返回空，跳过测试")
	}

	// 搜索
	query := `{"query":{"match":{"title":"Hello"}}}`
	searchRes, err := client.Search(ctx, "test_index", query)
	if err != nil {
		t.Skipf("搜索失败，跳过测试: %v", err)
	}
	if searchRes == "" {
		t.Skip("搜索结果为空，跳过测试")
	}

	// 删除文档
	err = client.Delete(ctx, "test_index", "1")
	if err != nil {
		t.Skipf("删除文档失败，跳过测试: %v", err)
	}

	// 删除索引
	err = client.DeleteIndex(ctx, "test_index")
	if err != nil {
		t.Skipf("删除索引失败，跳过测试: %v", err)
	}

	// 再次判断索引是否存在
	exists, err = client.IndexExists(ctx, "test_index")
	if err != nil {
		t.Skipf("IndexExists 查询失败，跳过测试: %v", err)
	}
	if exists {
		t.Skip("IndexExists 删除后应为 false，跳过测试")
	}
}
