package debug

import "context"

// Config debug config
type Config struct {
	Module []string `toml:"module" mapstructure:"module"` // debug模块, 可选值 mysql,http, + 自定义业务模块
	Mode   string   `toml:"mode" mapstructure:"mode"`     // debug方式, 可选值 log, console
}

type DebugFunc func(c context.Context, format string, v ...interface{})

func DefaultConfig() *Config {
	return &Config{
		Module: []string{},
		Mode:   "console",
	}
}
