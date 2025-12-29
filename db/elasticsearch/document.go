package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Index 文档写入/更新。body 可以是 io.Reader, []byte, string 或可被 json.Marshal 的结构体。
// refresh a string that controls when changes made by this request are made visible to search.
func (c *Client) Index(ctx context.Context, index string, id string, body interface{}, refresh string) error {
	reader, err := anaylzeBody(body)
	if err != nil {
		return err
	}

	opts := []func(*esapi.IndexRequest){
		c.ES.Index.WithDocumentID(id),
		c.ES.Index.WithContext(ctx),
	}
	if refresh != "" {
		opts = append(opts, c.ES.Index.WithRefresh(refresh))
	}

	res, err := c.ES.Index(index, reader, opts...)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.Status())
	}
	return nil
}

// Get 查询文档，并将结果解码到 out 中。
// 返回 (found bool, err error)。如果文档不存在 (404)，则 found=false, err=nil。
func (c *Client) Get(ctx context.Context, index string, id string, out interface{}) (found bool, err error) {
	res, err := c.ES.Get(index, id, c.ES.Get.WithContext(ctx))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, nil // Not Found is not an error, just not found.
	}
	if res.IsError() {
		return false, fmt.Errorf("get error from elasticsearch: %s", res.Status())
	}

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return true, fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return true, nil
}

// Delete 删除文档。
// 如果文档不存在 (404)，此方法不会返回错误。
// refresh a string that controls when changes made by this request are made visible to search.
func (c *Client) Delete(ctx context.Context, index string, id string, refresh string) error {
	opts := []func(*esapi.DeleteRequest){
		c.ES.Delete.WithContext(ctx),
	}
	if refresh != "" {
		opts = append(opts, c.ES.Delete.WithRefresh(refresh))
	}

	res, err := c.ES.Delete(index, id, opts...)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 404 Not Found means the document was already gone, which is not an error.
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.Status())
	}
	return nil
}

// Search 简单搜索。query 可以是 io.Reader, []byte, string 或可被 json.Marshal 的结构体。
// 结果将解码到 out 中。
func (c *Client) Search(ctx context.Context, index string, query interface{}, out interface{}) error {
	reader, err := anaylzeBody(query)
	if err != nil {
		return err
	}

	res, err := c.ES.Search(
		c.ES.Search.WithContext(ctx),
		c.ES.Search.WithIndex(index),
		c.ES.Search.WithBody(reader),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("search error: %s", res.Status())
	}

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode search result: %w", err)
		}
	}
	return nil
}

// DecodeResponse 是一个辅助函数，用于将响应体解码到给定的结构体中
func DecodeResponse(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// anaylzeBody 解析 body 为 io.Reader
func anaylzeBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch b := body.(type) {
	case io.Reader:
		return b, nil
	case []byte:
		return bytes.NewReader(b), nil
	case string:
		return bytes.NewReader([]byte(b)), nil
	default:
		// 尝试 JSON 序列化
		data, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		return bytes.NewReader(data), nil
	}
}