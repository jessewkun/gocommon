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
		MaxRequestBodySize:  1024 * 1024, // 1MB
		MaxResponseBodySize: 1024 * 1024, // 1MB
		SensitiveFields: []string{
			`(?i)password`,
			`(?i)token`,
			`(?i)secret`,
		},
		LogHeaders:    true,
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

		// 读取请求体
		var requestBody []byte
		if config.LogRequestBody && c.Request.Method == http.MethodPost {
			if c.Request.ContentLength > config.MaxRequestBodySize {
				requestBody = []byte("[请求体超过大小限制]")
			} else {
				requestBody, _ = io.ReadAll(c.Request.Body)
				// 恢复请求体
				c.Request.Body = io.NopCloser(strings.NewReader(string(requestBody)))
			}
		}

		// 处理请求
		c.Next()

		// 获取响应体
		var responseBody interface{}
		if config.LogResponseBody {
			responseBody, _ = c.Get(constant.CTX_API_OUTPUT)
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
		if config.LogRequestBody && len(requestBody) > 0 {
			fields["request_body"] = maskSensitiveData(string(requestBody), sensitiveRegexps)
		}

		// 添加响应体
		if config.LogResponseBody && responseBody != nil {
			fields["response"] = responseBody
		}

		// 添加响应大小
		fields["response_length"] = c.Writer.Size()

		// 根据状态码选择日志级别
		var logFunc func(c context.Context, tag string, msg string, field map[string]interface{})
		status := c.Writer.Status()
		switch {
		case status >= http.StatusInternalServerError:
			logFunc = logger.ErrorWithField
		case status >= http.StatusBadRequest:
			logFunc = logger.WarnWithField
		default:
			logFunc = logger.InfoWithField
		}

		// 记录日志
		logFunc(c.Request.Context(), TAG, "IOLOG", fields)
	}
}

// maskSensitiveData 对敏感数据进行脱敏处理
func maskSensitiveData(data string, patterns []*regexp.Regexp) string {
	for _, pattern := range patterns {
		data = pattern.ReplaceAllStringFunc(data, func(s string) string {
			return strings.Repeat("*", len(s))
		})
	}
	return data
}
