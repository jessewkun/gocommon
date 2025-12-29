package mongodb

import "github.com/jessewkun/gocommon/config"

type Config struct {
	Uris                   []string `mapstructure:"uris" json:"uris"`                                         // MongoDB 连接字符串列表
	MaxPoolSize            int      `mapstructure:"max_pool_size" json:"max_pool_size"`                       // 最大连接池大小，默认100
	MinPoolSize            int      `mapstructure:"min_pool_size" json:"min_pool_size"`                       // 最小连接池大小，默认5
	MaxConnIdleTime        int      `mapstructure:"max_conn_idle_time" json:"max_conn_idle_time"`             // 连接最大空闲时间，单位秒，默认300秒
	ConnectTimeout         int      `mapstructure:"connect_timeout" json:"connect_timeout"`                   // 连接超时时间，单位秒，默认10秒
	ServerSelectionTimeout int      `mapstructure:"server_selection_timeout" json:"server_selection_timeout"` // 服务器选择超时时间，单位秒，默认5秒
	SocketTimeout          int      `mapstructure:"socket_timeout" json:"socket_timeout"`                     // Socket超时时间，单位秒，默认30秒
	ReadPreference         string   `mapstructure:"read_preference" json:"read_preference"`                   // 读取偏好：primary, primaryPreferred, secondary, secondaryPreferred, nearest
	WriteConcern           string   `mapstructure:"write_concern" json:"write_concern"`                       // 写入关注：majority, 1, 0
	IsLog                  bool     `mapstructure:"is_log" json:"is_log"`                                     // 是否记录日志
	SlowThreshold          int      `mapstructure:"slow_threshold" json:"slow_threshold"`                     // 慢查询阈值，单位毫秒，默认500毫秒
}

var Cfgs = make(map[string]*Config)

func init() {
	config.Register("mongodb", &Cfgs)
	config.RegisterCallback("mongodb", Init, "config", "log")
}

// HealthStatus MongoDB 健康状态
type HealthStatus struct {
	Status    string `json:"status"`    // 状态：success/error
	Error     string `json:"error"`     // 错误信息
	Latency   int64  `json:"latency"`   // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"` // 检查时间戳
	MaxPool   int    `json:"max_pool"`  // 最大连接池大小
	InUse     int    `json:"in_use"`    // 正在使用连接数
	Idle      int    `json:"idle"`      // 空闲连接数 (MongoDB驱动不直接提供)
	Available int    `json:"available"` // 可用连接数
}
