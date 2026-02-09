// Package localcache 提供本地缓存功能
package localcache

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache"
	"github.com/jessewkun/gocommon/common"
)

// Cache 本地缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(key string) (interface{}, bool)
	// Set 设置缓存值
	Set(key string, value interface{}) error
	// SetWithTTL 设置缓存值并指定TTL
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
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

// Stats 缓存统计信息
type Stats struct {
	Hits        int64 `json:"hits"`        // 命中次数
	Misses      int64 `json:"misses"`      // 未命中次数
	Evictions   int64 `json:"evictions"`   // 淘汰次数
	Expirations int64 `json:"expirations"` // 过期次数
}

// HitRate 命中率
func (s Stats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// cacheItem 缓存项结构
type cacheItem struct {
	Value    interface{} `json:"value"`
	ExpireAt *time.Time  `json:"expire_at,omitempty"`
}

// BigCacheWrapper bigcache包装器
type BigCacheWrapper struct {
	cache *bigcache.BigCache
	stats Stats
}

// NewBigCache 创建新的bigcache缓存
func NewBigCache(config bigcache.Config) (Cache, error) {
	cache, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}

	wrapper := &BigCacheWrapper{
		cache: cache,
	}

	return wrapper, nil
}

// NewDefaultBigCache 创建默认配置的bigcache
func NewDefaultBigCache() (Cache, error) {
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.Verbose = false
	return NewBigCache(config)
}

// NewBigCacheWithSize 根据大小创建bigcache
func NewBigCacheWithSize(maxEntriesInWindow int) (Cache, error) {
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         10 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: maxEntriesInWindow,
		MaxEntrySize:       500,
		Verbose:            false,
		HardMaxCacheSize:   0,
		Logger:             nil,
	}
	return NewBigCache(config)
}

// Get 获取缓存值
func (bc *BigCacheWrapper) Get(key string) (interface{}, bool) {
	data, err := bc.cache.Get(key)
	if err != nil {
		atomic.AddInt64(&bc.stats.Misses, 1)
		return nil, false
	}

	// 解析缓存项
	var item cacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		atomic.AddInt64(&bc.stats.Misses, 1)
		return nil, false
	}

	// 检查是否为被删除的项（通过Delete方法设置的标记）
	if item.Value == nil {
		atomic.AddInt64(&bc.stats.Misses, 1)
		return nil, false
	}

	// 检查是否过期
	if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
		atomic.AddInt64(&bc.stats.Expirations, 1)
		atomic.AddInt64(&bc.stats.Misses, 1)
		// 同步删除，避免高并发下 goroutine 爆炸
		bc.deleteExpired(key)
		return nil, false
	}

	atomic.AddInt64(&bc.stats.Hits, 1)
	return item.Value, true
}

// Set 设置缓存值（无TTL）
func (bc *BigCacheWrapper) Set(key string, value interface{}) error {
	return bc.SetWithTTL(key, value, 0)
}

// SetWithTTL 设置缓存值并指定TTL
func (bc *BigCacheWrapper) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return common.NewCustomError(10001, fmt.Errorf("key cannot be empty"))
	}

	// 创建缓存项
	item := cacheItem{
		Value: value,
	}

	// 如果设置了TTL，计算过期时间
	if ttl > 0 {
		expireAt := time.Now().Add(ttl)
		item.ExpireAt = &expireAt
	}

	// 序列化缓存项
	data, err := json.Marshal(item)
	if err != nil {
		return common.NewCustomError(10002, err)
	}

	// 存储到bigcache
	err = bc.cache.Set(key, data)
	if err != nil {
		return common.NewCustomError(10003, err)
	}

	return nil
}

// Delete 删除缓存
func (bc *BigCacheWrapper) Delete(key string) bool {
	// bigcache不支持删除，我们通过设置一个value为nil的项来模拟删除
	emptyItem := cacheItem{Value: nil}
	data, err := json.Marshal(emptyItem)
	if err != nil {
		return false
	}

	err = bc.cache.Set(key, data)
	return err == nil
}

// ResetStats 重置统计信息
func (bc *BigCacheWrapper) ResetStats() {
	atomic.StoreInt64(&bc.stats.Hits, 0)
	atomic.StoreInt64(&bc.stats.Misses, 0)
	atomic.StoreInt64(&bc.stats.Evictions, 0)
	atomic.StoreInt64(&bc.stats.Expirations, 0)
}

// Size 获取缓存大小
func (bc *BigCacheWrapper) Size() int {
	return bc.cache.Len()
}

// Capacity 获取缓存容量
func (bc *BigCacheWrapper) Capacity() int {
	return bc.cache.Capacity()
}

// Stats 获取缓存统计信息
func (bc *BigCacheWrapper) Stats() Stats {
	return Stats{
		Hits:        atomic.LoadInt64(&bc.stats.Hits),
		Misses:      atomic.LoadInt64(&bc.stats.Misses),
		Evictions:   0, // bigcache不直接暴露Evictions计数
		Expirations: atomic.LoadInt64(&bc.stats.Expirations),
	}
}

// Close 关闭缓存
func (bc *BigCacheWrapper) Close() error {
	return bc.cache.Close()
}

// deleteExpired 异步删除过期项
func (bc *BigCacheWrapper) deleteExpired(key string) {
	bc.Delete(key)
}
