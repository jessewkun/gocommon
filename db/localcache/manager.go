package localcache

import (
	"sync"
)

// Manager 缓存管理器
type Manager struct {
	caches map[string]Cache
	mutex  sync.RWMutex
}

// NewManager 创建新的缓存管理器
func NewManager() *Manager {
	return &Manager{
		caches: make(map[string]Cache),
	}
}

// GetCache 获取或创建缓存
func (m *Manager) GetCache(name string, maxEntriesInWindow int) (Cache, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cache, exists := m.caches[name]; exists {
		return cache, nil
	}

	cache, err := NewBigCacheWithSize(maxEntriesInWindow)
	if err != nil {
		return nil, err
	}

	m.caches[name] = cache
	return cache, nil
}

// GetDefaultCache 获取或创建默认配置的缓存
func (m *Manager) GetDefaultCache(name string) (Cache, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cache, exists := m.caches[name]; exists {
		return cache, nil
	}

	cache, err := NewDefaultBigCache()
	if err != nil {
		return nil, err
	}

	m.caches[name] = cache
	return cache, nil
}

// GetTypedCache 是一个泛型辅助函数，用于从管理器中获取或创建类型安全的缓存。
// 注意：此函数为独立函数，而非Manager的方法，以正确支持泛型。
func GetTypedCache[T any](m *Manager, name string, maxEntriesInWindow int) (TypedCache[T], error) {
	// 复用GetCache的逻辑来获取底层的非类型化缓存
	cache, err := m.GetCache(name, maxEntriesInWindow)
	if err != nil {
		// 返回泛型零值
		var zero TypedCache[T]
		return zero, err
	}

	// 将获取到的Cache实例包装成类型安全的TypedCache
	return &typedCache[T]{cache: cache}, nil
}

// RemoveCache 移除缓存
func (m *Manager) RemoveCache(name string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cache, exists := m.caches[name]; exists {
		cache.Close()
		delete(m.caches, name)
		return true
	}
	return false
}

// ClearAll 关闭并移除所有缓存
func (m *Manager) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, cache := range m.caches {
		cache.Close()
		delete(m.caches, name)
	}
}

// ListCaches 列出所有缓存名称
func (m *Manager) ListCaches() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}

// GetCacheStats 获取指定缓存的统计信息
func (m *Manager) GetCacheStats(name string) (Stats, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if cache, exists := m.caches[name]; exists {
		return cache.Stats(), true
	}
	return Stats{}, false
}

// GetAllStats 获取所有缓存的统计信息
func (m *Manager) GetAllStats() map[string]Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]Stats)
	for name, cache := range m.caches {
		stats[name] = cache.Stats()
	}
	return stats
}
