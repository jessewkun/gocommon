package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jessewkun/gocommon/config"
	"github.com/spf13/viper"
)

// Config 是 http 模块的配置结构体
type Config struct {
	TransparentParameter []string `mapstructure:"transparent_parameter"`
	IsTraceLog           bool     `mapstructure:"is_trace_log"`
}

// Reload 重新加载 http 配置
func (c *Config) Reload(v *viper.Viper) {
	if err := v.UnmarshalKey("http", c); err != nil {
		fmt.Printf("failed to reload http config: %v\n", err)
	}
	fmt.Printf("http config reload success, config: %+v\n", c)
}

var Cfg = &Config{}

func init() {
	config.Register("http", Cfg)
}

type Option struct {
	Headers            map[string]string // 请求头
	Timeout            time.Duration     // 超时时间（秒）
	Retry              int               // 最大重试次数
	RetryWaitTime      time.Duration     // 重试等待时间
	RetryMaxWaitTime   time.Duration     // 最大重试等待时间
	RetryWith5xxStatus bool              // 是否对5xx状态码进行重试
	IsLog              bool              // 是否记录日志
}

func (o *Option) String() string {
	return fmt.Sprintf("Headers: %v, Timeout: %s, RetryCount: %d, RetryWaitTime: %s, RetryMaxWaitTime: %s, RetryWith5xxStatus: %v, IsLog: %v",
		o.Headers, o.Timeout, o.Retry, o.RetryWaitTime, o.RetryMaxWaitTime, o.RetryWith5xxStatus, o.IsLog)
}

// Response
type Response struct {
	StatusCode int             // http response status code
	Body       []byte          // http response body
	Header     http.Header     // http response header
	TraceInfo  resty.TraceInfo // http response trace info
}

func (h *Response) String() string {
	return fmt.Sprintf("Body: %s, Header: %v, StatusCode: %d, TraceInfo: %+v", h.Body, h.Header, h.StatusCode, h.TraceInfo)
}

// post request
type PostRequest struct {
	URL     string            // 请求地址
	Payload interface{}       // 请求数据
	Headers map[string]string // 请求头
	Timeout time.Duration     // 请求超时时间，如果为0则使用客户端默认超时时间
}

// upload request
type UploadRequest struct {
	URL       string            // 请求地址
	FileBytes []byte            // 文件字节
	Param     string            // 文件参数名
	FileName  string            // 文件名
	Data      map[string]string // 请求数据
	Headers   map[string]string // 请求头
	Timeout   time.Duration     // 请求超时时间，如果为0则使用客户端默认超时时间
}

// upload with file path request
type UploadWithFilePathRequest struct {
	URL      string            // 请求地址
	FileName string            // 文件名
	FilePath string            // 文件路径
	Param    string            // 文件参数名
	Data     map[string]string // 请求数据
	Headers  map[string]string // 请求头
	Timeout  time.Duration     // 请求超时时间，如果为0则使用客户端默认超时时间
}

// download request
type DownloadRequest struct {
	URL      string            // 请求地址
	FilePath string            // 文件路径
	Headers  map[string]string // 请求头
	Timeout  time.Duration     // 请求超时时间，如果为0则使用客户端默认超时时间
}

// get request
type GetRequest struct {
	URL     string            // 请求地址
	Headers map[string]string // 请求头
	Timeout time.Duration     // 请求超时时间，如果为0则使用客户端默认超时时间
}
