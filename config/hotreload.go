package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// HotReloadable 支持部分热更新的配置结构体接口
type HotReloadable interface {
	// Reload 当配置文件变化时调用。
	// 实现此接口的模块，应将 viper 的配置解析到自身，并返回可能发生的错误。
	Reload(v *viper.Viper) error
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

		// 为了保证原子性和线程安全，所有更新都在一个写锁内完成。
		// 这会短暂地阻塞所有配置的读取者，但能确保数据一致性。
		configMutex.Lock()
		defer configMutex.Unlock()

		var errs []string

		// 重新填充基础配置
		if err := v.Unmarshal(Cfg); err != nil {
			errs = append(errs, fmt.Sprintf("unmarshal base config: %v", err))
		}

		// 为支持热更新的模块重新加载配置
		modulesMutex.RLock()
		for key, cfgPtr := range modules {
			if reloader, ok := cfgPtr.(HotReloadable); ok {
				if err := reloader.Reload(v); err != nil {
					errs = append(errs, fmt.Sprintf("reload module '%s': %v", key, err))
				}
			}
		}
		modulesMutex.RUnlock()

		// 如果有任何错误，打印聚合日志。
		// 注意：在此策略下，配置可能已部分更新。这是一个权衡。
		// 更安全的策略是反解到副本，但会增加实现的复杂性。
		if len(errs) > 0 {
			combinedErr := errors.New(strings.Join(errs, "; "))
			fmt.Printf("Errors occurred during hot-reload (config might be inconsistent): %v\n", combinedErr)
			return
		}

		fmt.Println("hot reload success")
	})
}
