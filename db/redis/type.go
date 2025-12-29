package redis

import (
	"github.com/jessewkun/gocommon/config"
)

type Config struct {
	Addrs              []string `mapstructure:"addrs" json:"addrs"`                               // redis addrs ip:port
	Password           string   `mapstructure:"password" json:"password"`                         // redis password
	Db                 int      `mapstructure:"db" json:"db"`                                     // redis db
	IsCluster          bool     `mapstructure:"is_cluster" json:"is_cluster"`                     // 是否为集群模式
	IsLog              bool     `mapstructure:"is_log" json:"is_log"`                             // 是否记录日志
	PoolSize           int      `mapstructure:"pool_size" json:"pool_size"`                       // 连接池大小
	IdleTimeout        int      `mapstructure:"idle_timeout" json:"idle_timeout"`                 // 空闲连接超时时间，单位秒
	IdleCheckFrequency int      `mapstructure:"idle_check_frequency" json:"idle_check_frequency"` // 空闲连接检查频率，单位秒
	MinIdleConns       int      `mapstructure:"min_idle_conns" json:"min_idle_conns"`             // 最小空闲连接数
	MaxRetries         int      `mapstructure:"max_retries" json:"max_retries"`                   // 最大重试次数
	DialTimeout        int      `mapstructure:"dial_timeout" json:"dial_timeout"`                 // 连接超时时间，单位秒
	SlowThreshold      int      `mapstructure:"slow_threshold" json:"slow_threshold"`             // 慢查询阈值，单位毫秒
}

var (
	Cfgs           = make(map[string]*Config)
	defaultManager *Manager
)

func init() {
	config.Register("redis", &Cfgs)
	config.RegisterCallback("redis", Init, "config", "log")
}

// HealthStatus Redis健康状态
type HealthStatus struct {
	Status    string `json:"status"`    // 状态：success/error
	Error     string `json:"error"`     // 错误信息
	Latency   int64  `json:"latency"`   // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"` // 检查时间戳
}
