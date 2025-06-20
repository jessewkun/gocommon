package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// HotReloadable 支持部分热更新的配置结构体接口
type HotReloadable interface {
	// Reload 当配置文件变化时调用
	// 实现应该从 viper 实例安全地更新结构体的字段
	Reload(v *viper.Viper)
}

// hotReload 设置文件监听器，在配置文件变化时重新加载配置
func hotReload(v *viper.Viper) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)

		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config on hot-reload: %v\n", err)
			return
		}

		// 重新填充基础配置
		if err := v.Unmarshal(Cfg); err != nil {
			fmt.Printf("Error unmarshaling base config on hot-reload: %v\n", err)
		}

		// 为支持热更新的模块重新加载配置
		modulesMutex.RLock()
		defer modulesMutex.RUnlock()
		for _, cfgPtr := range modules {
			if reloader, ok := cfgPtr.(HotReloadable); ok {
				reloader.Reload(v)
			}
		}
		fmt.Println("hot reload success")
	})
}
