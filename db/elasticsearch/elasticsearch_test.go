package elasticsearch

import (
	"context"
	"testing"
)

var testCfg = Config{
	Addresses: []string{"http://127.0.0.1:9200"},
	Username:  "",
	Password:  "",
}

func TestElasticsearch_BasicFlow(t *testing.T) {
	client, err := NewClient(testCfg)
	if err != nil {
		t.Fatalf("ES连接失败: %v", err)
	}
	ctx := context.Background()

	// 健康检查
	hc := client.HealthCheck()
	if hc["elasticsearch"].Status != "success" {
		t.Errorf("健康检查失败: %+v", hc)
	}

	// 创建索引
	mapping := `{"mappings":{"properties":{"title":{"type":"text"},"content":{"type":"text"}}}}`
	err = client.CreateIndex(ctx, "test_index", mapping)
	if err != nil {
		t.Errorf("创建索引失败: %v", err)
	}

	// 判断索引是否存在
	exists, err := client.IndexExists(ctx, "test_index")
	if err != nil {
		t.Errorf("IndexExists 查询失败: %v", err)
	}
	if !exists {
		t.Error("IndexExists 应为 true")
	}

	// 获取索引 mapping
	m, err := client.GetIndexMapping(ctx, "test_index")
	if err != nil {
		t.Errorf("GetIndexMapping 失败: %v", err)
	}
	if m == "" {
		t.Error("GetIndexMapping 返回空")
	}

	// 写入文档
	doc := `{"title":"Hello ES","content":"Elasticsearch test content"}`
	err = client.Index(ctx, "test_index", "1", doc)
	if err != nil {
		t.Errorf("写入文档失败: %v", err)
	}

	// 查询文档
	res, err := client.Get(ctx, "test_index", "1")
	if err != nil {
		t.Errorf("查询文档失败: %v", err)
	}
	if res == "" {
		t.Error("查询文档返回空")
	}

	// 搜索
	query := `{"query":{"match":{"title":"Hello"}}}`
	searchRes, err := client.Search(ctx, "test_index", query)
	if err != nil {
		t.Errorf("搜索失败: %v", err)
	}
	if searchRes == "" {
		t.Error("搜索结果为空")
	}

	// 删除文档
	err = client.Delete(ctx, "test_index", "1")
	if err != nil {
		t.Errorf("删除文档失败: %v", err)
	}

	// 删除索引
	err = client.DeleteIndex(ctx, "test_index")
	if err != nil {
		t.Errorf("删除索引失败: %v", err)
	}

	// 再次判断索引是否存在
	exists, err = client.IndexExists(ctx, "test_index")
	if err != nil {
		t.Errorf("IndexExists 查询失败: %v", err)
	}
	if exists {
		t.Error("IndexExists 删除后应为 false")
	}
}
