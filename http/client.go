// Package http 提供 HTTP 客户端功能
package http

import (
	"github.com/go-resty/resty/v2"
	"github.com/jessewkun/gocommon/logger"
	"github.com/spf13/cast"
)

// 流式 buffer 默认值（字节）
const (
	DefaultStreamBufferInitial = 64 * 1024   // 64KB
	DefaultStreamBufferMax     = 1024 * 1024 // 1MB
)

type Client struct {
	client              *resty.Client
	streamBufferInitial int
	streamBufferMax     int
}

func NewClient(opt Option) *Client {
	client := resty.New()

	if opt.Timeout > 0 {
		client.SetTimeout(opt.Timeout)
	}

	// 设置 headers
	if opt.Headers != nil {
		client.SetHeaders(opt.Headers)
	}

	// 设置重试策略
	if opt.Retry > 0 {
		client.SetRetryCount(opt.Retry)
		if opt.RetryWaitTime > 0 {
			client.SetRetryWaitTime(opt.RetryWaitTime)
		}
		if opt.RetryMaxWaitTime > 0 {
			client.SetRetryMaxWaitTime(opt.RetryMaxWaitTime)
		}
		// 根据配置决定是否对5xx状态码进行重试
		if opt.RetryWith5xxStatus {
			client.AddRetryCondition(func(r *resty.Response, err error) bool {
				return r.StatusCode() >= 500 && r.StatusCode() < 600
			})
		}
	}

	// 日志逻辑现在尊重模块的配置，可以被 Option 覆盖
	isLog := Cfg.IsLog
	if opt.IsLog != nil {
		isLog = *opt.IsLog
	}

	if isLog {
		client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			ctx := r.Request.Context()
			logger.InfoWithField(ctx, TAG, "client request", map[string]interface{}{
				"url":       r.Request.URL,
				"respData":  r,
				"traceInfo": r.Request.TraceInfo(),
				"header":    r.Request.Header,
			})
			return nil
		})
	}

	// 透传参数钩子
	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		ctx := r.Context()
		// 每次都从配置中读取最新的透传参数，支持热更新
		for _, parameter := range Cfg.TransparentParameter {
			if value := ctx.Value(parameter); value != nil {
				r.SetHeader(parameter, cast.ToString(value))
			}
		}
		return nil
	})

	streamBufInit := opt.StreamBufferInitial
	streamBufMax := opt.StreamBufferMax
	if streamBufInit <= 0 {
		streamBufInit = DefaultStreamBufferInitial
	}
	if streamBufMax <= 0 {
		streamBufMax = DefaultStreamBufferMax
	}

	return &Client{
		client:              client,
		streamBufferInitial: streamBufInit,
		streamBufferMax:     streamBufMax,
	}
}
