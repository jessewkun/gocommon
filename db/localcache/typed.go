package localcache

import (
	"time"
)

// TypedCache 类型安全的缓存接口
type TypedCache[T any] interface {
	// Get 获取缓存值
	Get(key string) (T, bool)
	// Set 设置缓存值
	Set(key string, value T) error
	// SetWithTTL 设置缓存值并指定TTL
	SetWithTTL(key string, value T, ttl time.Duration) error
	// Delete 删除缓存
	Delete(key string) bool
	// Clear 清空缓存
	Clear()
	// Size 获取缓存大小
	Size() int
	// Capacity 获取缓存容量
	Capacity() int
	// Stats 获取缓存统计信息
	Stats() Stats
	// Close 关闭缓存
	Close() error
}

// typedCache 类型安全缓存的实现
type typedCache[T any] struct {
	cache Cache
}

// NewTypedBigCache 创建新的类型安全bigcache
func NewTypedBigCache[T any](maxEntriesInWindow int) (TypedCache[T], error) {
	cache, err := NewBigCacheWithSize(maxEntriesInWindow)
	if err != nil {
		return nil, err
	}

	return &typedCache[T]{
		cache: cache,
	}, nil
}

// NewDefaultTypedBigCache 创建默认配置的类型安全bigcache
func NewDefaultTypedBigCache[T any]() (TypedCache[T], error) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		return nil, err
	}

	return &typedCache[T]{
		cache: cache,
	}, nil
}

// Get 获取缓存值
func (tc *typedCache[T]) Get(key string) (T, bool) {
	value, exists := tc.cache.Get(key)
	if !exists {
		var zero T
		return zero, false
	}

	if typedValue, ok := value.(T); ok {
		return typedValue, true
	}

	var zero T
	return zero, false
}

// Set 设置缓存值
func (tc *typedCache[T]) Set(key string, value T) error {
	return tc.cache.Set(key, value)
}

// SetWithTTL 设置缓存值并指定TTL
func (tc *typedCache[T]) SetWithTTL(key string, value T, ttl time.Duration) error {
	return tc.cache.SetWithTTL(key, value, ttl)
}

// Delete 删除缓存
func (tc *typedCache[T]) Delete(key string) bool {
	return tc.cache.Delete(key)
}

// Clear 清空缓存
func (tc *typedCache[T]) Clear() {
	tc.cache.Clear()
}

// Size 获取缓存大小
func (tc *typedCache[T]) Size() int {
	return tc.cache.Size()
}

// Capacity 获取缓存容量
func (tc *typedCache[T]) Capacity() int {
	return tc.cache.Capacity()
}

// Stats 获取缓存统计信息
func (tc *typedCache[T]) Stats() Stats {
	return tc.cache.Stats()
}

// Close 关闭缓存
func (tc *typedCache[T]) Close() error {
	return tc.cache.Close()
}
