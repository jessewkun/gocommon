package debug

import (
	"context"
	"fmt"

	"github.com/jessewkun/gocommon/config"
	"github.com/spf13/viper"
)

type Config struct {
	Module []string `toml:"module" mapstructure:"module"` // debug模块, 可选值 mysql,http, + 自定义业务模块
	Mode   string   `toml:"mode" mapstructure:"mode"`     // debug方式, 可选值 log, console
}

func (c *Config) Reload(v *viper.Viper) {
	if err := v.UnmarshalKey("debug", c); err != nil {
		fmt.Printf("failed to reload debug config: %v\n", err)
		return
	}
	fmt.Printf("debug config reload success, config: %+v\n", c)
}

var Cfg = DefaultConfig()

func init() {
	config.Register("debug", Cfg)
}

type DebugFunc func(c context.Context, format string, v ...interface{})

func DefaultConfig() *Config {
	return &Config{
		Module: []string{},
		Mode:   "console",
	}
}
