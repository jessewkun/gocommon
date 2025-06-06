package config

import (
	"github.com/jessewkun/gocommon/alarm"
	"github.com/jessewkun/gocommon/cache"
	"github.com/jessewkun/gocommon/db"
	"github.com/jessewkun/gocommon/debug"
	"github.com/jessewkun/gocommon/logger"
	"github.com/spf13/viper"
)

// 项目通用配置
type BaseConfig struct {
	Mode   string                   `toml:"mode" mapstructure:"mode"`     // 运行模式, debug 开发, release 生产, test 测试
	Port   string                   `toml:"port" mapstructure:"port"`     // 服务端口, 默认 :8000
	Domain string                   `toml:"domain" mapstructure:"domain"` // 服务域名, 默认 http://localhost:8000
	Debug  *debug.Config            `toml:"debug" mapstructure:"debug"`   // 调试配置
	Log    *logger.Config           `toml:"log" mapstructure:"log"`       // 日志配置
	Mysql  map[string]*db.Config    `toml:"mysql" mapstructure:"mysql"`   // mysql 配置
	Redis  map[string]*cache.Config `toml:"redis" mapstructure:"redis"`   // redis 配置
	Alarm  *alarm.Config            `toml:"alarm" mapstructure:"alarm"`   // 报警配置
}

// DefaultConfig 返回默认配置
func DefaultConfig() *BaseConfig {
	return &BaseConfig{
		Mode:   "debug",
		Port:   ":8000",
		Domain: "http://localhost:8000",
		Debug:  debug.DefaultConfig(),
		Log:    logger.DefaultConfig(),
		Mysql:  make(map[string]*db.Config),
		Redis:  make(map[string]*cache.Config),
		Alarm:  alarm.DefaultConfig(),
	}
}

// 项目配置
var Cfg BaseConfig

// LoadConfig 加载配置
func LoadConfig(configPath string) (*BaseConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
