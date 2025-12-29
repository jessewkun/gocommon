package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"


)

// CreateIndex 创建索引。mapping 可以是 io.Reader, []byte, string 或可被 json.Marshal 的结构体。
// 此方法会处理并关闭响应体，调用者无需关心。
func (c *Client) CreateIndex(ctx context.Context, index string, mapping interface{}) error {
	reader, err := anaylzeBody(mapping)
	if err != nil {
		return err
	}
	res, err := c.ES.Indices.Create(index, c.ES.Indices.Create.WithBody(reader), c.ES.Indices.Create.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.Status())
	}
	return nil
}

// DeleteIndex 删除索引。
// 如果索引不存在 (404)，此方法不会返回错误。
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	res, err := c.ES.Indices.Delete([]string{index}, c.ES.Indices.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete index error: %s", res.Status())
	}
	return nil
}

// IndexExists 判断索引是否存在
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := c.ES.Indices.Exists([]string{index}, c.ES.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, err
	}
	defer res.Body.Close() // Exists is a HEAD request, body is empty, but good practice.

	if res.StatusCode == 200 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("index exists check error: %s", res.Status())
}

// GetIndexMapping 获取索引 mapping，并将结果解码到 out 中。
func (c *Client) GetIndexMapping(ctx context.Context, index string, out interface{}) error {
	res, err := c.ES.Indices.GetMapping(c.ES.Indices.GetMapping.WithIndex(index), c.ES.Indices.GetMapping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("get mapping error: %s", res.Status())
	}

	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}