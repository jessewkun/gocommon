// Package config 提供配置管理和热重载功能
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

	// 主要针对需要在加载配置后做进一步初始化的模块，比如 mysql 的建连
	callbacks      = make(map[string]func() error)
	callbacksMutex sync.RWMutex
)

// Register 配置注册，用于注册模块配置，key 为模块名，cfgPtr 为模块配置结构体指针
func Register(key string, cfgPtr interface{}) {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	if _, exists := modules[key]; exists {
		panic(fmt.Sprintf("config: module '%s' is already registered", key))
	}
	modules[key] = cfgPtr
}

// RegisterCallback 注册模块初始化回调函数，用于在加载配置后初始化模块
func RegisterCallback(key string, fn func() error) {
	callbacksMutex.Lock()
	defer callbacksMutex.Unlock()

	if _, exists := callbacks[key]; exists {
		panic(fmt.Sprintf("config: module '%s' is already registered", key))
	}
	callbacks[key] = fn
}

// BaseConfig 基础配置
type BaseConfig struct {
	Mode    string `mapstructure:"mode" json:"mode"`         // 运行模式, debug 开发, release 生产, test 测试
	Port    string `mapstructure:"port" json:"port"`         // 服务端口, 默认 :8000
	AppName string `mapstructure:"app_name" json:"app_name"` // 服务标题, 默认 "Service"
	Domain  string `mapstructure:"domain" json:"domain"`     // 服务域名, 默认 http://localhost:8000
}

// Cfg is the global instance for base configuration.
var Cfg = &BaseConfig{
	Mode:    "debug",
	Port:    ":8000",
	AppName: "Service",
	Domain:  "http://localhost:8000",
}

// Init 初始化配置，从给定路径加载配置，填充基础配置 (Cfg)，并填充所有注册的模块配置
// 注意：该函数会自动调用所有注册的回调函数，所以请确保回调函数是幂等的
func Init(configPath string) (*BaseConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(Cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base config: %w", err)
	}

	allSettings := v.AllSettings()
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()

	// 初始化各个模块的配置
	for key, cfgPtr := range modules {
		if _, ok := allSettings[key]; !ok {
			fmt.Printf("config: key '%s' not found in config file, module will use default values\n", key)
		}
		if err := v.UnmarshalKey(key, cfgPtr); err != nil {
			return nil, fmt.Errorf("config: failed to unmarshal key '%s': %v", key, err)
		}
	}

	if err := callModuleCallback(); err != nil {
		return nil, err
	}

	go hotReload(v)

	return Cfg, nil
}

// callModuleCallback 调用模块的初始化回调函数
func callModuleCallback() error {
	callbacksMutex.RLock()
	defer callbacksMutex.RUnlock()

	// 确保 log 和 alarm 优先初始化，其他模块依赖 log 和 alarm
	// 备注：模块的注册是依赖的 init 函数，不能依赖模块的注册顺序来初始化模块，会有意想不到的错误
	if fn, ok := callbacks["alarm"]; ok {
		if err := fn(); err != nil {
			return err
		}
		delete(callbacks, "alarm")
	}

	if fn, ok := callbacks["log"]; ok {
		if err := fn(); err != nil {
			return err
		}
		delete(callbacks, "log")
	}

	// 注意如果有前后依赖关系的模块，需要确保先初始化依赖的模块
	for _, fn := range callbacks {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
