package logger

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jessewkun/gocommon/utils"
)

// Config 日志配置
type Config struct {
	Path                 string   `toml:"path" mapstructure:"path"`                                   // ⽇志⽂件路径
	Closed               bool     `toml:"closed" mapstructure:"closed"`                               // 是否关闭日志，注意该配置是全局配置，一旦关闭则所有日志都不会输出
	MaxSize              int      `toml:"max_size" mapstructure:"max_size"`                           // 单位为MB,默认为100MB
	MaxAge               int      `toml:"max_age" mapstructure:"max_age"`                             // 文件最多保存多少天
	MaxBackup            int      `toml:"max_backup" mapstructure:"max_backup"`                       // 保留多少个备份
	TransparentParameter []string `toml:"transparent_parameter" mapstructure:"transparent_parameter"` // 透传参数，继承上下文中的参数
	AlarmLevel           string   `toml:"alarm_level" mapstructure:"alarm_level"`                     // 报警级别, warn 警告, error 错误
}

// Validate 验证配置是否合法
func (c *Config) Validate() error {
	if c.Path == "" {
		return errors.New("log path cannot be empty")
	}

	// 确保日志目录存在
	dir := filepath.Dir(c.Path)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	if c.MaxSize <= 0 {
		return errors.New("max size must be greater than 0")
	}

	if c.MaxAge <= 0 {
		return errors.New("max age must be greater than 0")
	}

	if c.MaxBackup <= 0 {
		return errors.New("max backup must be greater than 0")
	}

	// 验证报警级别
	if _, ok := alarmLevelMap[c.AlarmLevel]; !ok {
		return fmt.Errorf("invalid alarm level: %s", c.AlarmLevel)
	}

	return nil
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
