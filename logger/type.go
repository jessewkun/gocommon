package logger

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jessewkun/gocommon/config"
	"github.com/jessewkun/gocommon/utils"
)

// Config 日志配置
// 高并发系统: 增大 MaxSize，减小 MaxAge，适当增加 MaxBackup
// 低频率系统: 减小 MaxSize，增大 MaxAge，减少 MaxBackup
// 磁盘空间紧张: 减小所有三个参数
// 需要长期保留: 增大 MaxAge 和 MaxBackup
type Config struct {
	Path                 string   `mapstructure:"path" json:"path"`                                   // ⽇志⽂件路径
	Closed               bool     `mapstructure:"closed" json:"closed"`                               // 是否关闭日志，注意该配置是全局配置，一旦关闭则所有日志都不会输出
	MaxSize              int      `mapstructure:"max_size" json:"max_size"`                           // 单位为MB,默认为100MB
	MaxAge               int      `mapstructure:"max_age" json:"max_age"`                             // 文件最多保存多少天
	MaxBackup            int      `mapstructure:"max_backup" json:"max_backup"`                       // 保留多少个备份
	TransparentParameter []string `mapstructure:"transparent_parameter" json:"transparent_parameter"` // 透传参数，继承上下文中的参数
	AlarmLevel           string   `mapstructure:"alarm_level" json:"alarm_level"`                     // 报警级别, warn 警告, error 错误
}

var Cfg = DefaultConfig()

func init() {
	config.Register("log", Cfg)
	config.RegisterCallback("log", Init)
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
	if c.AlarmLevel != "" {
		if _, ok := alarmLevelMap[c.AlarmLevel]; !ok {
			return fmt.Errorf("invalid alarm level: %s", c.AlarmLevel)
		}
	}

	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Path:                 "",
		Closed:               true,
		MaxSize:              100,
		MaxAge:               30,
		MaxBackup:            10,
		TransparentParameter: []string{},
		AlarmLevel:           "warn",
	}
}
