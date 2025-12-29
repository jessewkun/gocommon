// Package config provides a flexible configuration management solution with support for multiple configuration formats, module-based configuration registration, dependency management, hot reloading, cross-module dependency injection, and basic configuration management.
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

	// 模块初始化回调函数及其依赖
	callbacks      = make(map[string]ModuleCallback)
	callbacksMutex sync.RWMutex

	// 注入器函数，用于处理跨模块依赖注入
	injectorFn    func() error
	injectorMutex sync.Mutex

	// configMutex 保护所有配置对象在热更新时的并发写安全
	configMutex sync.RWMutex
)

// ModuleCallback 封装了模块初始化回调函数及其依赖
type ModuleCallback struct {
	Fn           func() error
	Dependencies []string
}

// Register 配置注册，用于注册模块配置，key 为模块名，cfgPtr 为模块配置结构体指针
func Register(key string, cfgPtr interface{}) {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	if _, exists := modules[key]; exists {
		panic(fmt.Sprintf("config: module '%s' is already registered", key))
	}
	modules[key] = cfgPtr
}

// RegisterCallback 注册模块初始化回调函数
//
// fn func() error 回调函数
//
// dependencies []string 依赖的模块
func RegisterCallback(key string, fn func() error, dependencies ...string) {
	callbacksMutex.Lock()
	defer callbacksMutex.Unlock()

	if _, exists := callbacks[key]; exists {
		panic(fmt.Sprintf("config: module '%s' is already registered", key))
	}
	callbacks[key] = ModuleCallback{
		Fn:           fn,
		Dependencies: dependencies,
	}
}

// RegisterInjector 注册一个跨模块依赖注入函数
// 该函数将在所有模块回调执行完毕后被调用
func RegisterInjector(fn func() error) {
	injectorMutex.Lock()
	defer injectorMutex.Unlock()
	if injectorFn != nil {
		panic("config: injector function is already registered")
	}
	injectorFn = fn
}

// BaseConfig 基础配置
type BaseConfig struct {
	Mode    string `mapstructure:"mode" json:"mode"`         // 运行模式, debug 开发, release 生产, test 测试
	Port    string `mapstructure:"port" json:"port"`         // 服务端口, 默认 8000
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

	// 加载配置到结构体，加写锁保证初始化过程的原子性
	configMutex.Lock()
	if err := v.Unmarshal(Cfg); err != nil {
		configMutex.Unlock()
		return nil, fmt.Errorf("failed to unmarshal base config: %w", err)
	}

	allSettings := v.AllSettings()
	modulesMutex.RLock()

	// 初始化各个模块的配置
	for key, cfgPtr := range modules {
		if _, ok := allSettings[key]; !ok {
			fmt.Printf("config: key '%s' not found in config file, module will use default values\n", key)
		}
		if err := v.UnmarshalKey(key, cfgPtr); err != nil {
			modulesMutex.RUnlock()
			configMutex.Unlock()
			return nil, fmt.Errorf("config: failed to unmarshal key '%s': %v", key, err)
		}
	}
	modulesMutex.RUnlock()
	configMutex.Unlock()

	if err := callModuleCallback(); err != nil {
		return nil, err
	}

	// 调用已注册的跨模块依赖注入函数，由业务侧注入这个函数
	injectorMutex.Lock()
	if injectorFn != nil {
		if err := injectorFn(); err != nil {
			injectorMutex.Unlock()
			return nil, fmt.Errorf("registered injector function failed: %w", err)
		}
	}
	injectorMutex.Unlock()

	go hotReload(v)

	return Cfg, nil
}

// callModuleCallback 调用模块的初始化回调函数，按依赖关系排序
func callModuleCallback() error {
	callbacksMutex.RLock()
	defer callbacksMutex.RUnlock()

	// 构建依赖图和入度
	graph := make(map[string][]string) // adj list: module -> [modules that depend on it]
	inDegree := make(map[string]int)   // module -> count of its dependencies

	// 预填充所有模块和其依赖，以确保图中包含所有可能节点
	for module, cb := range callbacks {
		inDegree[module] = 0 // 所有有回调的模块都初始化入度为0
		// 确保所有依赖项也作为图节点存在
		for _, dep := range cb.Dependencies {
			if _, ok := inDegree[dep]; !ok {
				inDegree[dep] = 0
			}
		}
	}

	// 填充图的边和入度
	for module := range callbacks {
		inDegree[module] = 0 // 先为所有有回调的模块初始化入度
	}
	for module, cb := range callbacks {
		for _, dep := range cb.Dependencies {
			// 关键修复：只在依赖项本身也有回调时，才建立图中的边和增加入度
			if _, depHasCallback := callbacks[dep]; depHasCallback {
				graph[dep] = append(graph[dep], module) // dep 是 module 的前置条件
				inDegree[module]++
			}
			// 如果依赖项（如 'http' 或 'config'）没有回调，我们忽略它，
			// 因为它的配置加载已在 `Init` 函数早期完成，该依赖已满足。
		}
	}

	// 找到所有入度为0的节点 (没有未满足依赖的模块)
	var queue []string
	for module := range callbacks {
		if inDegree[module] == 0 {
			queue = append(queue, module)
		}
	}

	// 拓扑排序的结果
	var sortedModules []string
	executedModuleCount := 0 // 计数已执行的回调模块数量

	for len(queue) > 0 {
		// Pop module from queue
		module := queue[0]
		queue = queue[1:]

		// 执行当前模块的回调
		cb := callbacks[module]
		if cb.Fn != nil { // 确保有回调函数才执行
			if err := cb.Fn(); err != nil {
				// 延续旧逻辑：log模块失败中断，其他模块只打印错误
				if module == "log" {
					return fmt.Errorf("config: callback for '%s' failed, cannot continue: %w", module, err)
				}
				fmt.Printf("config: callback for module '%s' failed, but continuing: %v\n", module, err)
			}
		}
		sortedModules = append(sortedModules, module)
		executedModuleCount++

		// 遍历所有依赖当前模块的模块
		for _, neighbor := range graph[module] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 { // 当依赖都满足时
				// 只有有回调函数的才加入队列
				if _, hasCallback := callbacks[neighbor]; hasCallback {
					queue = append(queue, neighbor)
				}
			}
		}
	}

	// 检查是否存在循环依赖
	if executedModuleCount != len(callbacks) {
		unexecutedModules := make([]string, 0)
		for module := range callbacks {
			found := false
			for _, sorted := range sortedModules {
				if module == sorted {
					found = true
					break
				}
			}
			if !found {
				unexecutedModules = append(unexecutedModules, module)
			}
		}
		if len(unexecutedModules) > 0 {
			return fmt.Errorf("config: circular dependency detected involving modules: %v", unexecutedModules)
		}
		// 理论上不会走到这里，如果 executedModuleCount != len(callbacks) 且没有循环依赖，那就是图构建逻辑有问题
		return fmt.Errorf("config: internal graph error, executed callbacks count %d != total callbacks count %d", executedModuleCount, len(callbacks))
	}

	return nil
}
