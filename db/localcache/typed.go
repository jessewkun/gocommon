package localcache

import (
	"encoding/json"
	"errors"
	"fmt"
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
	// ResetStats 重置统计信息。注意：此方法不会清空缓存中的数据。
	ResetStats()
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

	// 尝试直接类型断言，这适用于非复杂类型
	if typedValue, ok := value.(T); ok {
		return typedValue, true
	}

	// 如果断言失败，则尝试通过JSON解码
	decoded, err := decodeTypedValue[T](value)
	if err != nil {
		var zero T
		return zero, false
	}

	return decoded, true
}

// Set 设置缓存值
func (tc *typedCache[T]) Set(key string, value T) error {
	payload, err := encodeTypedValue(value)
	if err != nil {
		return err
	}
	return tc.cache.Set(key, payload)
}

// SetWithTTL 设置缓存值并指定TTL
func (tc *typedCache[T]) SetWithTTL(key string, value T, ttl time.Duration) error {
	payload, err := encodeTypedValue(value)
	if err != nil {
		return err
	}
	return tc.cache.SetWithTTL(key, payload, ttl)
}

// Delete 删除缓存
func (tc *typedCache[T]) Delete(key string) bool {
	return tc.cache.Delete(key)
}

// ResetStats 重置统计信息
func (tc *typedCache[T]) ResetStats() {
	tc.cache.ResetStats()
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

// decodeTypedValue 通过 JSON 编码/解码将 interface{} 转换为具体类型
func decodeTypedValue[T any](value interface{}) (T, error) {
	var zero T

	// 因为Set时，复杂类型被转为string(json)，所以这里优先处理string
	if str, ok := value.(string); ok {
		if str == "" {
			return zero, errors.New("empty cached value string")
		}
		return decodeJSONValue[T]([]byte(str))
	}

	// 其次处理map[string]interface{}，这是json.Unmarshal后的默认复杂类型
	data, err := json.Marshal(value)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal cached value: %w", err)
	}
	return decodeJSONValue[T](data)
}

// encodeTypedValue 对结构体等复杂类型做一次 JSON 序列化，避免在 Get 时再次编码
func encodeTypedValue[T any](value T) (interface{}, error) {
	// 对于基础类型，直接返回，避免不必要的json序列化
	switch v := any(value).(type) {
	case nil,
		string,
		[]byte,
		bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return v, nil
	}

	// 对于其他所有复杂类型（structs, maps, slices），进行JSON序列化
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	// 返回string而不是[]byte, 因为json.Unmarshal(interface{})会把[]byte当成base64编码的string
	return string(data), nil
}

func decodeJSONValue[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}
	return result, nil
}
