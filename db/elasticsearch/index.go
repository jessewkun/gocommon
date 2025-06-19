package elasticsearch

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, index string, mapping string) error {
	res, err := c.ES.Indices.Create(index, c.ES.Indices.Create.WithBody(strings.NewReader(mapping)), c.ES.Indices.Create.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}
	return nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	res, err := c.ES.Indices.Delete([]string{index}, c.ES.Indices.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.String())
	}
	return nil
}

// IndexExists 判断索引是否存在
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := c.ES.Indices.Exists([]string{index}, c.ES.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("index exists check error: %s", res.String())
}

// GetIndexMapping 获取索引 mapping
func (c *Client) GetIndexMapping(ctx context.Context, index string) (string, error) {
	res, err := c.ES.Indices.GetMapping(c.ES.Indices.GetMapping.WithIndex(index), c.ES.Indices.GetMapping.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", fmt.Errorf("get mapping error: %s", res.String())
	}
	var b strings.Builder
	_, err = io.Copy(&b, res.Body)
	return b.String(), err
}
