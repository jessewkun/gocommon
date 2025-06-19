package elasticsearch

import "github.com/elastic/go-elasticsearch/v8"

type Client struct {
	ES *elasticsearch.Client
}

// Config 用于初始化 ES 客户端
// Example: Config{Addresses: []string{"http://localhost:9200"}}
type Config struct {
	Addresses []string `toml:"addresses" mapstructure:"addresses"`
	Username  string   `toml:"username" mapstructure:"username"`
	Password  string   `toml:"password" mapstructure:"password"`
}

// HealthStatus ES健康状态
// 连接数等字段可留空或为0，主要关注状态、错误、延迟、时间戳
// 便于与 mysql/mongodb 健康检查结构统一
type HealthStatus struct {
	Status    string `json:"status"`    // 状态：success/error
	Error     string `json:"error"`     // 错误信息
	Latency   int64  `json:"latency"`   // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"` // 检查时间戳
}
