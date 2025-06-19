package elasticsearch

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Index 文档写入/更新
func (c *Client) Index(ctx context.Context, index string, id string, body string) error {
	res, err := c.ES.Index(index, strings.NewReader(body), c.ES.Index.WithDocumentID(id), c.ES.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}
	return nil
}

// Get 查询文档
func (c *Client) Get(ctx context.Context, index string, id string) (string, error) {
	res, err := c.ES.Get(index, id, c.ES.Get.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", fmt.Errorf("get error: %s", res.String())
	}
	var b strings.Builder
	_, err = io.Copy(&b, res.Body)
	return b.String(), err
}

// Delete 删除文档
func (c *Client) Delete(ctx context.Context, index string, id string) error {
	res, err := c.ES.Delete(index, id, c.ES.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("delete error: %s", res.String())
	}
	return nil
}

// Search 简单搜索
func (c *Client) Search(ctx context.Context, index string, query string) (string, error) {
	res, err := c.ES.Search(
		c.ES.Search.WithContext(ctx),
		c.ES.Search.WithIndex(index),
		c.ES.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", fmt.Errorf("search error: %s", res.String())
	}
	var b strings.Builder
	_, err = io.Copy(&b, res.Body)
	return b.String(), err
}
