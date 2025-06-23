package alarm

import (
	"fmt"

	"github.com/jessewkun/gocommon/config"
	"github.com/spf13/viper"
)

type Config struct {
	BarkIds []string `toml:"bark_ids" mapstructure:"bark_ids"` // Bark 设备 ID 列表
	Timeout int      `toml:"timeout" mapstructure:"timeout"`   // 请求超时时间（秒）
}

// Reload 重新加载 alarm 配置.
// alarm模块的所有配置项都被认为是安全的，可以进行热更新.
func (c *Config) Reload(v *viper.Viper) {
	if err := v.UnmarshalKey("alarm", c); err != nil {
		fmt.Printf("failed to reload alarm config: %v\n", err)
		return
	}
	fmt.Printf("alarm config reload success, config: %+v\n", c)
}

var Cfg = DefaultConfig()

func init() {
	config.Register("alarm", Cfg)
	config.RegisterCallback("alarm", Init)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BarkIds: []string{},
		Timeout: 5,
	}
}
