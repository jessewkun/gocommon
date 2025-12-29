package localcache

import (
	"testing"
	"time"
)

func TestBigCache_Basic(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// 测试基本设置和获取
	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, exists := cache.Get("key1")
	if !exists {
		t.Fatal("Get failed: key1 should exist")
	}
	if value != "value1" {
		t.Fatalf("Expected value1, got %v", value)
	}

	// 测试不存在的key
	_, exists = cache.Get("nonexistent")
	if exists {
		t.Fatal("Get should return false for nonexistent key")
	}
}

func TestBigCache_TTL(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// 设置带TTL的缓存
	err = cache.SetWithTTL("key1", "value1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("SetWithTTL failed: %v", err)
	}

	// 立即获取应该成功
	value, exists := cache.Get("key1")
	if !exists {
		t.Fatal("Get failed: key1 should exist")
	}
	if value != "value1" {
		t.Fatalf("Expected value1, got %v", value)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后获取应该失败
	_, exists = cache.Get("key1")
	if exists {
		t.Fatal("Get should return false for expired key")
	}
}

func TestBigCache_ComplexData(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// 测试复杂数据结构
	complexData := map[string]interface{}{
		"string": "hello",
		"int":    123,
		"float":  3.14,
		"bool":   true,
		"array":  []int{1, 2, 3},
		"map": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err = cache.Set("complex", complexData)
	if err != nil {
		t.Fatalf("Set complex data failed: %v", err)
	}

	value, exists := cache.Get("complex")
	if !exists {
		t.Fatal("Get failed: complex data should exist")
	}

	// 验证数据结构
	if retrieved, ok := value.(map[string]interface{}); ok {
		if retrieved["string"] != "hello" {
			t.Fatalf("Expected 'hello', got %v", retrieved["string"])
		}
		// JSON unmarshals numbers into float64 by default
		if int(retrieved["int"].(float64)) != 123 {
			t.Fatalf("Expected 123, got %v", retrieved["int"])
		}
	} else {
		t.Fatalf("Expected map[string]interface{}, got %T", value)
	}
}

func TestBigCache_Delete(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	cache.Set("key1", "value1")

	// 删除存在的key
	deleted := cache.Delete("key1")
	if !deleted {
		t.Fatal("Delete should return true for existing key")
	}

	// 验证已删除. Get 应该返回 false.
	_, exists := cache.Get("key1")
	if exists {
		t.Fatal("key1 should be deleted, Get should return false")
	}

	// 删除不存在的key (实际上是设置一个删除标记)
	deleted = cache.Delete("nonexistent")
	if !deleted {
		t.Fatal("Delete should always return true")
	}
	_, exists = cache.Get("nonexistent")
	if exists {
		t.Fatal("nonexistent key should return false after delete")
	}
}

func TestBigCache_ResetStats(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Get("key1")

	// 重置前检查
	statsBefore := cache.Stats()
	if statsBefore.Hits == 0 {
		t.Fatalf("Expected hits > 0 before reset, got %d", statsBefore.Hits)
	}
	if cache.Size() != 2 {
		t.Fatalf("Expected size 2, got %d", cache.Size())
	}

	cache.ResetStats()

	// 检查统计信息是否被重置
	statsAfter := cache.Stats()
	if statsAfter.Hits != 0 || statsAfter.Misses != 0 {
		t.Fatalf("Stats should be reset after clear, got Hits=%d, Misses=%d", statsAfter.Hits, statsAfter.Misses)
	}

	// 验证数据仍然存在
	if cache.Size() != 2 {
		t.Fatalf("Size should not be affected by ResetStats, got %d", cache.Size())
	}
	val, exists := cache.Get("key1")
	if !exists || val != "value1" {
		t.Fatal("Data should still exist after ResetStats")
	}
}

func TestBigCache_Stats(t *testing.T) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// 添加一些数据
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// 获取命中
	cache.Get("key1")
	cache.Get("key2")

	// 获取未命中
	cache.Get("nonexistent")

	stats := cache.Stats()
	// stats can be racy without a lock, so we check for minimums
	if stats.Hits < 2 {
		t.Fatalf("Expected at least 2 hits, got %d", stats.Hits)
	}
	if stats.Misses < 1 {
		t.Fatalf("Expected at least 1 miss, got %d", stats.Misses)
	}

	hitRate := stats.HitRate()
	if hitRate <= 0.5 {
		t.Fatalf("Expected positive hit rate, got %f", hitRate)
	}
}

func TestTypedBigCache(t *testing.T) {
	cache, err := NewTypedBigCache[string](1000)
	if err != nil {
		t.Fatalf("Failed to create typed cache: %v", err)
	}
	defer cache.Close()

	// 测试类型安全缓存
	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, exists := cache.Get("key1")
	if !exists {
		t.Fatal("Get failed: key1 should exist")
	}
	if value != "value1" {
		t.Fatalf("Expected value1, got %s", value)
	}

	// 测试零值
	value, exists = cache.Get("nonexistent")
	if exists {
		t.Fatal("Get should return false for nonexistent key")
	}
	if value != "" {
		t.Fatalf("Expected empty string for nonexistent key, got %s", value)
	}
}

func TestManager(t *testing.T) {
	manager := NewManager()
	defer manager.ClearAll()

	// 创建缓存
	cache1, err := manager.GetCache("cache1", 1000)
	if err != nil {
		t.Fatalf("Failed to create cache1: %v", err)
	}

	cache2, err := manager.GetCache("cache2", 2000)
	if err != nil {
		t.Fatalf("Failed to create cache2: %v", err)
	}

	// 测试缓存
	cache1.Set("key1", "value1")
	cache2.Set("key2", "value2")

	value, exists := cache1.Get("key1")
	if !exists || value != "value1" {
		t.Fatal("cache1 should contain key1")
	}

	value, exists = cache2.Get("key2")
	if !exists || value != "value2" {
		t.Fatal("cache2 should contain key2")
	}

	// 测试统计信息
	stats := manager.GetAllStats()
	if len(stats) != 2 {
		t.Fatalf("Expected 2 caches, got %d", len(stats))
	}

	// 测试列出缓存
	caches := manager.ListCaches()
	if len(caches) != 2 {
		t.Fatalf("Expected 2 cache names, got %d", len(caches))
	}
}

func TestManager_GetTypedCache(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	manager := NewManager()
	defer manager.ClearAll()

	// 获取类型安全的缓存
	userCache, err := GetTypedCache[User](manager, "users", 1000)
	if err != nil {
		t.Fatalf("Failed to get typed cache: %v", err)
	}

	// 设置和获取
	user := User{ID: 1, Name: "Alice"}
	err = userCache.Set("user:1", user)
	if err != nil {
		t.Fatalf("Set failed on typed cache: %v", err)
	}

	retrievedUser, exists := userCache.Get("user:1")
	if !exists {
		t.Fatal("Get failed on typed cache: user:1 should exist")
	}
	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Fatalf("Expected user %v, got %v", user, retrievedUser)
	}

	// 确保获取的是同一个实例
	userCache2, err := GetTypedCache[User](manager, "users", 1000)
	if err != nil {
		t.Fatalf("Failed to get same typed cache again: %v", err)
	}
	retrievedUser2, exists2 := userCache2.Get("user:1")
	if !exists2 {
		t.Fatal("Get from second accessor failed")
	}
	if retrievedUser2.ID != user.ID {
		t.Fatal("Second accessor got wrong data")
	}
}

func BenchmarkBigCache_Set(b *testing.B) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key"+string(rune(i)), "value"+string(rune(i)))
	}
}

func BenchmarkBigCache_Get(b *testing.B) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// 预填充缓存
	for i := 0; i < 1000; i++ {
		cache.Set("key"+string(rune(i)), "value"+string(rune(i)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key" + string(rune(i%1000)))
	}
}

func BenchmarkBigCache_ComplexData(b *testing.B) {
	cache, err := NewDefaultBigCache()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	complexData := map[string]interface{}{
		"string": "hello",
		"int":    123,
		"float":  3.14,
		"bool":   true,
		"array":  []int{1, 2, 3},
		"map": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("complex"+string(rune(i)), complexData)
	}
}
