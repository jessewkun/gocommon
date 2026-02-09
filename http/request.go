package http

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/spf13/cast"
)

// BuildQuery http build query
func BuildQuery(data map[string]interface{}) string {
	var uri url.URL

	q := uri.Query()
	for k, v := range data {
		q.Add(k, cast.ToString(v))
	}
	return q.Encode()
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, req RequestGet) (*Response, error) {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx)

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Get(req.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       resp.Body(),
		Header:     resp.Header(),
		StatusCode: resp.StatusCode(),
		TraceInfo:  resp.Request.TraceInfo(),
	}, nil
}

// Post 发送POST请求
func (c *Client) Post(ctx context.Context, req RequestPost) (*Response, error) {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx).SetBody(req.Payload)

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Post(req.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       resp.Body(),
		Header:     resp.Header(),
		StatusCode: resp.StatusCode(),
		TraceInfo:  resp.Request.TraceInfo(),
	}, nil
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, req RequestUpload) (respData *Response, err error) {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx).
		SetFileReader(req.Param, req.FileName, bytes.NewReader(req.FileBytes)).
		SetFormData(req.Data)

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Post(req.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       resp.Body(),
		Header:     resp.Header(),
		StatusCode: resp.StatusCode(),
		TraceInfo:  resp.Request.TraceInfo(),
	}, nil
}

// UploadWithFilePath 上传文件
func (c *Client) UploadWithFilePath(ctx context.Context, req RequestUploadWithFilePath) (respData *Response, err error) {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx).
		SetFile(req.Param, req.FilePath).
		SetFormData(req.Data)

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Post(req.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       resp.Body(),
		Header:     resp.Header(),
		StatusCode: resp.StatusCode(),
		TraceInfo:  resp.Request.TraceInfo(),
	}, nil
}

// Download 下载文件
func (c *Client) Download(ctx context.Context, req RequestDownload) (respData *Response, err error) {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx).SetOutput(req.FilePath)

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Get(req.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		Header:     resp.Header(),
		StatusCode: resp.StatusCode(),
		TraceInfo:  resp.Request.TraceInfo(),
	}, nil
}

// PostStream 发送 POST 请求并以流式处理响应
func (c *Client) PostStream(ctx context.Context, req RequestPost, callback func(line []byte) error) error {
	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := c.client.R().SetContext(ctx).
		SetBody(req.Payload).
		SetDoNotParseResponse(true) // 核心：告诉 resty 不要自动解析响应体

	// 设置请求头
	if req.Headers != nil {
		request.SetHeaders(req.Headers)
	}

	resp, err := request.Post(req.URL)
	if err != nil {
		return fmt.Errorf("发送流式请求失败: %w", err)
	}
	rawBody := resp.RawResponse.Body
	defer rawBody.Close()

	if resp.IsError() {
		// 尝试读取错误信息
		bodyBytes, readErr := io.ReadAll(rawBody)
		if readErr != nil {
			return fmt.Errorf("API 返回错误状态 %d, 且读取错误响应体失败: %w", resp.StatusCode(), readErr)
		}
		return fmt.Errorf("API 返回错误状态 %d: %s", resp.StatusCode(), string(bodyBytes))
	}

	scanner := bufio.NewScanner(rawBody)
	buf := make([]byte, 0, c.streamBufferInitial)
	scanner.Buffer(buf, c.streamBufferMax)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := callback(line); err != nil {
			return fmt.Errorf("流式处理回调函数出错: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应体时出错: %w", err)
	}

	return nil
}
