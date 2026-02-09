package llm

import (
	"fmt"
	"sync"
)

// Factory 根据配置创建 Provider，config 由各实现方自行解析（如 map 或具体类型）
type Factory func(config interface{}) (Provider, error)

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register 注册 Provider 工厂，name 如 "openrouter"、"openai"
// 业务通过 NewProvider(name, config) 获取 Provider，无需引用具体子包
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[name]; exists {
		panic("llm: provider already registered: " + name)
	}
	registry[name] = factory
}

// NewProvider 按名称创建 Provider，config 传给对应工厂（如 map 或实现方规定的配置结构）
// 业务只需 import llm，例如：p, err := llm.NewProvider("openrouter", map[string]interface{}{"api_url": "...", "api_key": "..."})
func NewProvider(name string, config interface{}) (Provider, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("llm: unknown provider %q", name)
	}
	return factory(config)
}

// ProviderNames 返回已注册的 Provider 名称列表
func ProviderNames() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	return names
}
