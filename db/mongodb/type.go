package mongodb

// Config MongoDB 配置
type Config struct {
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
type HealthStatus struct {
	Status    string `json:"status"`    // 状态：success/error
	Error     string `json:"error"`     // 错误信息
	Latency   int64  `json:"latency"`   // 延迟，单位毫秒
	Timestamp int64  `json:"timestamp"` // 检查时间戳
	MaxPool   int    `json:"max_pool"`  // 最大连接池大小
	InUse     int    `json:"in_use"`    // 正在使用连接数
	Idle      int    `json:"idle"`      // 空闲连接数
	Available int    `json:"available"` // 可用连接数
}
