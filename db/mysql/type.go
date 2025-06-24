package mysql

import (
	"sync"
	"time"

	"github.com/jessewkun/gocommon/config"
	"gorm.io/gorm"
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

var Cfgs = make(map[string]*Config)

const TAG = "MYSQL"

type Connections struct {
	mu    sync.RWMutex
	conns map[string]*gorm.DB
}

var connList = &Connections{
	conns: make(map[string]*gorm.DB),
}

func init() {
	config.Register("mysql", &Cfgs)
	config.RegisterCallback("mysql", Init)
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
