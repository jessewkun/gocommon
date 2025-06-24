package http

import (
	"bytes"
	"context"
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

// Get
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

// Post
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

// Upload
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

// UploadWithFilePath
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

// Download
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
