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

// setTransparentParameter set transparent parameter
func (hc *Client) setTransparentParameter(c context.Context) *Client {
	if len(hc.TransparentParameter) > 0 {
		for _, parameter := range hc.TransparentParameter {
			if value := c.Value(parameter); value != nil {
				hc.Client.SetHeader(parameter, cast.ToString(value))
			}
		}
	}
	return hc
}

// Get
func (hc *Client) Get(ctx context.Context, req GetRequest) (*Response, error) {
	hc.setTransparentParameter(ctx)

	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := hc.Client.R().
		SetContext(ctx)

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
func (hc *Client) Post(ctx context.Context, req PostRequest) (*Response, error) {
	hc.setTransparentParameter(ctx)

	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := hc.Client.R().
		SetContext(ctx).
		SetBody(req.Payload)

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
func (hc *Client) Upload(ctx context.Context, req UploadRequest) (respData *Response, err error) {
	hc.setTransparentParameter(ctx)

	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := hc.Client.R().
		SetContext(ctx).
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
func (hc *Client) UploadWithFilePath(ctx context.Context, req UploadWithFilePathRequest) (respData *Response, err error) {
	hc.setTransparentParameter(ctx)

	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := hc.Client.R().
		SetContext(ctx).
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
func (hc *Client) Download(ctx context.Context, req DownloadRequest) (respData *Response, err error) {
	hc.setTransparentParameter(ctx)

	// 设置请求超时
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	request := hc.Client.R().
		SetContext(ctx).
		SetOutput(req.FilePath)

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
