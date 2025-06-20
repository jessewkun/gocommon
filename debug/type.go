package debug

import (
	"context"
	"fmt"

	"github.com/jessewkun/gocommon/config"
	"github.com/spf13/viper"
)

// Config debug config
type Config struct {
	Module []string `toml:"module" mapstructure:"module"` // debug模块, 可选值 mysql,http, + 自定义业务模块
	Mode   string   `toml:"mode" mapstructure:"mode"`     // debug方式, 可选值 log, console
}

// Reload 安全地重新加载 debug 配置
func (c *Config) Reload(v *viper.Viper) {
	if v.IsSet("debug.module") {
		c.Module = v.GetStringSlice("debug.module")
		fmt.Println("Debug module reloaded:", c.Module)
	}
	if v.IsSet("debug.mode") {
		c.Mode = v.GetString("debug.mode")
		fmt.Println("Debug mode reloaded:", c.Mode)
	}
	fmt.Printf("debug config reload success\n")
}

// Cfg 是 debug 模块的配置实例
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
