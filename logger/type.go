package logger

// Config 日志配置
type Config struct {
	Path                 string   `toml:"path" mapstructure:"path"`                                   // ⽇志⽂件路径
	Closed               bool     `toml:"closed" mapstructure:"closed"`                               // 是否关闭日志
	MaxSize              int      `toml:"max_size" mapstructure:"max_size"`                           // 单位为MB,默认为100MB
	MaxAge               int      `toml:"max_age" mapstructure:"max_age"`                             // 文件最多保存多少天
	MaxBackup            int      `toml:"max_backup" mapstructure:"max_backup"`                       // 保留多少个备份
	TransparentParameter []string `toml:"transparent_parameter" mapstructure:"transparent_parameter"` // 透传参数，继承上下文中的参数
	AlarmLevel           string   `toml:"alarm_level" mapstructure:"alarm_level"`                     // 报警级别, warn 警告, error 错误
}

func DefaultConfig() *Config {
	return &Config{
		Path:                 "",
		Closed:               false,
		MaxSize:              100,
		MaxAge:               30,
		MaxBackup:            10,
		TransparentParameter: []string{},
		AlarmLevel:           "warn",
	}
}

// 报警级别映射
var alarmLevelMap = map[string][]string{
	"debug": {"debug", "info", "warn", "error", "fatal", "panic"},
	"info":  {"info", "warn", "error", "fatal", "panic"},
	"warn":  {"warn", "error", "fatal", "panic"},
	"error": {"error", "fatal", "panic"},
	"fatal": {"fatal", "panic"},
	"panic": {"panic"},
}
