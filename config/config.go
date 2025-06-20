package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	// 全局配置注册表
	modules      = make(map[string]interface{})
	modulesMutex sync.RWMutex
)

// 配置注册，用于注册模块配置，key 为模块名，cfgPtr 为模块配置结构体指针
func Register(key string, cfgPtr interface{}) {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	if _, exists := modules[key]; exists {
		panic(fmt.Sprintf("config: module '%s' is already registered", key))
	}
	modules[key] = cfgPtr
}

// BaseConfig 基础配置
type BaseConfig struct {
	Mode   string `mapstructure:"mode"`   // 运行模式, debug 开发, release 生产, test 测试
	Port   string `mapstructure:"port"`   // 服务端口, 默认 :8000
	Domain string `mapstructure:"domain"` // 服务域名, 默认 http://localhost:8000
}

// Cfg is the global instance for base configuration.
var Cfg = &BaseConfig{
	Mode:   "debug",
	Port:   ":8000",
	Domain: "http://localhost:8000",
}

// LoadConfig 加载配置，从给定路径加载配置，填充基础配置 (Cfg)，并填充所有注册的模块配置
func LoadConfig(configPath string) (*BaseConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 将整个配置解码到一个临时 map 中以检查键
	allSettings := v.AllSettings()

	// 填充基础配置
	if err := v.Unmarshal(Cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base config: %w", err)
	}

	modulesMutex.RLock()
	defer modulesMutex.RUnlock()
	for key, cfgPtr := range modules {
		if _, ok := allSettings[key]; !ok {
			panic(fmt.Sprintf("config: key '%s' not found in config file, module will use default values\n", key))
		}
		if err := v.UnmarshalKey(key, cfgPtr); err != nil {
			panic(fmt.Sprintf("config: failed to unmarshal key '%s': %v\n", key, err))
		}
	}

	hotReload(v)

	return Cfg, nil
}
