package mysql

import (
	"time"

	"gorm.io/gorm/logger"
)

// Config mysql config
type Config struct {
	Dsn                       []string `toml:"dsn" mapstructure:"dsn"`                                                     // 数据源
	MaxConn                   int      `toml:"max_conn" mapstructure:"max_conn"`                                           // 最大连接数
	MaxIdleConn               int      `toml:"max_idle_conn" mapstructure:"max_idle_conn"`                                 // 最大空闲连接数
	ConnMaxLife               int      `toml:"conn_max_life" mapstructure:"conn_max_life"`                                 // 连接最长持续时间， 默认1小时，单位秒
	SlowThreshold             int      `toml:"slow_threshold" mapstructure:"slow_threshold"`                               // 慢查询阈值，单位毫秒，默认500毫秒
	IgnoreRecordNotFoundError bool     `toml:"ignore_record_not_found_error" mapstructure:"ignore_record_not_found_error"` // 是否忽略记录未找到错误
	IsLog                     bool     `toml:"is_log" mapstructure:"is_log"`                                               // 是否记录日志  日志级别为info
}

// HealthStatus MySQL健康状态
type HealthStatus struct {
	Status    string `json:"status"`     // 状态：success/error
	Error     string `json:"error"`      // 错误信息
	Latency   int64  `json:"latency"`    // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"`  // 检查时间戳
	MaxOpen   int    `json:"max_open"`   // 最大连接数
	Open      int    `json:"open"`       // 当前打开连接数
	InUse     int    `json:"in_use"`     // 正在使用连接数
	Idle      int    `json:"idle"`       // 空闲连接数
	WaitCount int64  `json:"wait_count"` // 等待连接数
	WaitTime  int64  `json:"wait_time"`  // 等待时间，单位纳秒
}

type mysqlLogger struct {
	SlowThreshold             time.Duration   // 慢查询阈值
	LogLevel                  logger.LogLevel // 日志级别
	IgnoreRecordNotFoundError bool            // 是否忽略记录未找到错误
}

// MongoConfig MongoDB 配置
type MongoConfig struct {
	Uris                   []string `toml:"uris" mapstructure:"uris"`                                         // MongoDB 连接字符串列表
	MaxPoolSize            int      `toml:"max_pool_size" mapstructure:"max_pool_size"`                       // 最大连接池大小，默认100
	MinPoolSize            int      `toml:"min_pool_size" mapstructure:"min_pool_size"`                       // 最小连接池大小，默认5
	MaxConnIdleTime        int      `toml:"max_conn_idle_time" mapstructure:"max_conn_idle_time"`             // 连接最大空闲时间，单位秒，默认300秒
	ConnectTimeout         int      `toml:"connect_timeout" mapstructure:"connect_timeout"`                   // 连接超时时间，单位秒，默认10秒
	ServerSelectionTimeout int      `toml:"server_selection_timeout" mapstructure:"server_selection_timeout"` // 服务器选择超时时间，单位秒，默认5秒
	SocketTimeout          int      `toml:"socket_timeout" mapstructure:"socket_timeout"`                     // Socket超时时间，单位秒，默认30秒
	ReadPreference         string   `toml:"read_preference" mapstructure:"read_preference"`                   // 读取偏好：primary, primaryPreferred, secondary, secondaryPreferred, nearest
	WriteConcern           string   `toml:"write_concern" mapstructure:"write_concern"`                       // 写入关注：majority, 1, 0
	IsLog                  bool     `toml:"is_log" mapstructure:"is_log"`                                     // 是否记录日志
}

// MongoHealthStatus MongoDB 健康状态
type MongoHealthStatus struct {
	Status    string `json:"status"`    // 状态：success/error
	Error     string `json:"error"`     // 错误信息
	Latency   int64  `json:"latency"`   // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"` // 检查时间戳
	MaxPool   int    `json:"max_pool"`  // 最大连接池大小
	InUse     int    `json:"in_use"`    // 正在使用连接数
	Idle      int    `json:"idle"`      // 空闲连接数
	Available int    `json:"available"` // 可用连接数
}
