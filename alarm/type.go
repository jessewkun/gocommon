package alarm

// Config 报警配置
type Config struct {
	BarkIds []string `toml:"bark_ids" mapstructure:"bark_ids"` // Bark 设备 ID 列表
	Timeout int      `toml:"timeout" mapstructure:"timeout"`   // 请求超时时间（秒）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BarkIds: []string{},
		Timeout: 5,
	}
}
