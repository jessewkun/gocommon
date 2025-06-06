package cache

// Config redis config
type Config struct {
	Addrs              []string `toml:"addrs" mapstructure:"addrs"`                               // redis addrs ip:port
	Password           string   `toml:"password" mapstructure:"password"`                         // redis password
	Db                 int      `toml:"db" mapstructure:"db"`                                     // redis db
	IsLog              bool     `toml:"is_log" mapstructure:"is_log"`                             // 是否记录日志
	PoolSize           int      `toml:"pool_size" mapstructure:"pool_size"`                       // 连接池大小
	IdleTimeout        int      `toml:"idle_timeout" mapstructure:"idle_timeout"`                 // 空闲连接超时时间，单位秒
	IdleCheckFrequency int      `toml:"idle_check_frequency" mapstructure:"idle_check_frequency"` // 空闲连接检查频率，单位秒
	MinIdleConns       int      `toml:"min_idle_conns" mapstructure:"min_idle_conns"`             // 最小空闲连接数
	MaxRetries         int      `toml:"max_retries" mapstructure:"max_retries"`                   // 最大重试次数
	DialTimeout        int      `toml:"dial_timeout" mapstructure:"dial_timeout"`                 // 连接超时时间，单位秒
}
