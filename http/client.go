package http

import (
	"github.com/go-resty/resty/v2"
	"github.com/jessewkun/gocommon/logger"
)

type Client struct {
	Client               *resty.Client
	TransparentParameter []string `toml:"transparent_parameter" mapstructure:"transparent_parameter"` // 透传参数，继承上下文中的参数
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
	isLog := Cfg.IsTraceLog
	if opt.IsLog {
		isLog = true
	}

	if isLog {
		client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			ctx := r.Request.Context()
			logger.InfoWithField(ctx, "HTTP", "client request", map[string]interface{}{
				"client":    c,
				"url":       r.Request.URL,
				"respData":  r,
				"traceInfo": r.Request.TraceInfo(),
				"header":    r.Request.Header,
			})
			return nil
		})
	}

	return &Client{
		Client:               client,
		TransparentParameter: Cfg.TransparentParameter,
	}
}
