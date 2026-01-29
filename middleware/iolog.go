package middleware

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jessewkun/gocommon/constant"
	"github.com/jessewkun/gocommon/logger"

	"github.com/gin-gonic/gin"
)

const (
	// SkipBodyLogKey 用于在 context 中标记是否跳过请求体或响应体日志的键
	SkipBodyLogKey = "skip_body_log"
)

type SkipBodyType int

const (
	SkipAllBody      SkipBodyType = 0
	SkipRequestBody  SkipBodyType = 1
	SkipResponseBody SkipBodyType = 2
)

// IOLogConfig 配置选项
type IOLogConfig struct {
	// 是否记录请求体
	LogRequestBody bool
	// 是否记录响应体
	LogResponseBody bool
	// 请求体大小限制（字节）
	MaxRequestBodySize int64
	// 响应体大小限制（字节）
	MaxResponseBodySize int64
	// 需要脱敏的字段（正则表达式）
	SensitiveFields []string
	// 是否记录请求头
	LogHeaders bool
	// 是否记录查询参数
	LogQuery bool
	// 是否记录路径参数
	LogPath bool
	// 是否记录客户端信息
	LogClientInfo bool
}

// DefaultIOLogConfig 返回默认配置
func DefaultIOLogConfig() *IOLogConfig {
	return &IOLogConfig{
		LogRequestBody:      true,
		LogResponseBody:     true,
		MaxRequestBodySize:  100 * 1024, // 100KB
		MaxResponseBodySize: 100 * 1024, // 100KB
		SensitiveFields: []string{
			`(?i)password`,
			`(?i)token`,
			`(?i)secret`,
		},
		LogHeaders:    false,
		LogQuery:      true,
		LogPath:       true,
		LogClientInfo: true,
	}
}

// IOLog 返回一个 IOLog 中间件
func IOLog(config *IOLogConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultIOLogConfig()
	}

	// 编译敏感字段正则表达式
	sensitiveRegexps := make([]*regexp.Regexp, len(config.SensitiveFields))
	for i, pattern := range config.SensitiveFields {
		sensitiveRegexps[i] = regexp.MustCompile(pattern)
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体（需要在 c.Next() 之前读取，因为请求体只能读取一次）
		// 如果后续设置了 skipRequestBody，则不记录到日志中
		var requestBody []byte
		if config.LogRequestBody && c.Request.Method == http.MethodPost {
			if c.Request.ContentLength > config.MaxRequestBodySize {
				requestBody = []byte("[请求体超过大小限制]")
			} else {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(strings.NewReader(string(requestBody)))
			}
		}

		// 处理请求（此时 SkipBodyLog 等中间件会执行并设置标记）
		c.Next()

		// 检查是否跳过请求体或响应体日志（在 c.Next() 之后检查，此时标记已经设置）
		skipRequestBody := false
		skipResponseBody := false
		if skipType, exists := c.Get(SkipBodyLogKey); exists {
			if bodyType, ok := skipType.(SkipBodyType); ok {
				switch bodyType {
				case SkipAllBody:
					skipRequestBody = true
					skipResponseBody = true
				case SkipRequestBody:
					skipRequestBody = true
				case SkipResponseBody:
					skipResponseBody = true
				}
			}
		}

		// 获取响应大小
		responseSize := c.Writer.Size()

		// 获取响应体
		var responseBody interface{}
		if config.LogResponseBody && !skipResponseBody {
			if config.MaxResponseBodySize > 0 && int64(responseSize) > config.MaxResponseBodySize {
				responseBody = "[响应体超过大小限制]"
			} else {
				responseBody, _ = c.Get(string(constant.CtxAPIOutput))
			}
		}

		// 构建日志字段
		fields := make(map[string]interface{})

		// 添加基本信息
		fields["duration"] = time.Since(start)
		fields["method"] = c.Request.Method
		fields["status"] = c.Writer.Status()

		// 添加请求信息
		if config.LogPath {
			fields["path"] = c.Request.URL.Path
		}
		if config.LogQuery {
			fields["query"] = c.Request.URL.RawQuery
		}
		if config.LogHeaders {
			fields["headers"] = c.Request.Header
		}
		if config.LogClientInfo {
			fields["client_ip"] = c.ClientIP()
			fields["user_agent"] = c.Request.UserAgent()
		}

		// 添加请求体
		if config.LogRequestBody && len(requestBody) > 0 && !skipRequestBody {
			fields["request_body"] = maskSensitiveData(string(requestBody), sensitiveRegexps)
		}

		// 添加响应体
		if config.LogResponseBody && responseBody != nil {
			fields["response"] = responseBody
		}

		// 添加响应大小
		fields["response_length"] = responseSize

		// 根据状态码选择日志级别
		var logFunc func(c context.Context, tag string, msg string, field map[string]interface{})
		status := c.Writer.Status()
		switch {
		case status >= http.StatusInternalServerError:
			logFunc = logger.ErrorWithField
		// case status >= http.StatusBadRequest:
		// 	logFunc = logger.WarnWithField
		default:
			logFunc = logger.InfoWithField
		}

		// 记录日志
		logFunc(c.Request.Context(), "IOLOG", http.StatusText(status), fields)
	}
}

// SkipBodyLog 返回一个中间件函数，用于标记跳过请求体或响应体日志记录，
// 一般要求都记录，但是部分接口响应体比较大，可以跳过记录，例如文件下载、文件上传等接口。
// 可以在单个路由或路由组中使用，例如：
//
//	router.GET("/api/export", SkipBodyLog(SkipResponseBody), exportHandler)
//	或
//	exportGroup := router.Group("/api/export", SkipBodyLog(SkipResponseBody))
func SkipBodyLog(skipBodyType SkipBodyType) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(SkipBodyLogKey, skipBodyType)
		c.Next()
	}
}

// maskSensitiveData 对敏感数据进行脱敏处理
func maskSensitiveData(data string, patterns []*regexp.Regexp) string {
	if len(patterns) == 0 {
		return data
	}

	for _, pattern := range patterns {
		data = pattern.ReplaceAllStringFunc(data, func(s string) string {
			return strings.Repeat("*", len(s))
		})
	}
	return data
}
