package debug

import (
	"context"
	"sync"

	"github.com/jessewkun/gocommon/config"
	"github.com/jessewkun/gocommon/logger"
	"github.com/spf13/viper"
)

// DebugFunc is the type for a debug logging function.
type DebugFunc func(c context.Context, format string, v ...interface{})

// Config defines the structure for debug configuration.
type Config struct {
	Module []string `mapstructure:"module" json:"module"` // Modules to enable debug for, e.g., mysql, http, or custom business modules.
	Mode   string   `mapstructure:"mode" json:"mode"`     // Debug output mode: "log" or "console".
}

// debugger is the internal struct that manages debug state and configuration.
type debugger struct {
	mu             sync.RWMutex
	Config         Config              `mapstructure:",squash"` // Exported to be filled by viper on initial load.
	enabledModules map[string]struct{} // For O(1) lookups.
}

// global default debugger instance.
var defaultDebugger = newDebugger()

func init() {
	// Register the debugger instance itself. On Init, viper will populate the exported `Config` field.
	config.Register("debug", defaultDebugger)
	// Register a callback to build the initial map after the config has been loaded.
	config.RegisterCallback("debug", defaultDebugger.buildInitialMap, "config", "log")
}

func newDebugger() *debugger {
	return &debugger{
		Config:         *defaultConfig(),
		enabledModules: make(map[string]struct{}),
	}
}

// buildInitialMap is called via callback after the initial config load. It now returns an error to match the interface.
func (d *debugger) buildInitialMap() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	newEnabledModules := make(map[string]struct{})
	for _, moduleName := range d.Config.Module {
		newEnabledModules[moduleName] = struct{}{}
	}
	d.enabledModules = newEnabledModules
	logger.Info(context.Background(), "DEBUG", "debug initial map built, config: %+v", d.Config)
	return nil
}

// Reload 重新加载 debug 配置.
// debug 模块的所有配置项都被认为是安全的，可以进行热更新.
func (d *debugger) Reload(v *viper.Viper) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 先将配置重置为默认值，以处理配置项被删除的情况。
	d.Config = *defaultConfig()

	// Unmarshal a "debug" key into the debugger instance itself.
	// The "squash" tag, added previously, handles the nesting.
	if err := v.UnmarshalKey("debug", d); err != nil {
		logger.ErrorWithMsg(context.Background(), "DEBUG", "failed to reload debug config: %v", err)
		return err
	}

	// 根据新加载的配置重建性能优化的 map。
	newEnabledModules := make(map[string]struct{})
	for _, moduleName := range d.Config.Module {
		newEnabledModules[moduleName] = struct{}{}
	}
	d.enabledModules = newEnabledModules

	logger.Info(context.Background(), "DEBUG", "debug config reloaded successfully, new config: %+v", d.Config)
	return nil
}

// isDebug checks if a specific module is enabled for debugging in a thread-safe manner.
func (d *debugger) isDebug(flag string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.enabledModules[flag]
	return ok
}

// defaultConfig returns a new default Config instance.
func defaultConfig() *Config {
	return &Config{
		Module: []string{},
		Mode:   "console",
	}
}
