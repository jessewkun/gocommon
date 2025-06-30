package localcache

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache"
)

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 创建默认配置的bigcache
	cache, err := NewDefaultBigCache()
	if err != nil {
		fmt.Printf("创建缓存失败: %v\n", err)
		return
	}
	defer cache.Close()

	// 设置缓存值
	err = cache.Set("user:123", map[string]interface{}{
		"id":   123,
		"name": "张三",
		"age":  25,
	})
	if err != nil {
		fmt.Printf("设置缓存失败: %v\n", err)
		return
	}

	// 获取缓存值
	value, exists := cache.Get("user:123")
	if exists {
		fmt.Printf("获取到用户信息: %v\n", value)
	} else {
		fmt.Println("用户信息不存在")
	}

	// 设置带TTL的缓存
	err = cache.SetWithTTL("session:abc", "session_data", 30*time.Minute)
	if err != nil {
		fmt.Printf("设置会话缓存失败: %v\n", err)
		return
	}

	// 删除缓存
	deleted := cache.Delete("user:123")
	if deleted {
		fmt.Println("用户信息已删除")
	}

	// 获取缓存统计信息
	stats := cache.Stats()
	fmt.Printf("缓存统计: 命中=%d, 未命中=%d, 命中率=%.2f%%\n",
		stats.Hits, stats.Misses, stats.HitRate()*100)
}

// ExampleTypedCache 类型安全缓存示例
func ExampleTypedCache() {
	// 创建字符串类型缓存
	stringCache, err := NewTypedBigCache[string](1000)
	if err != nil {
		fmt.Printf("创建字符串缓存失败: %v\n", err)
		return
	}
	defer stringCache.Close()

	// 创建用户结构体类型缓存
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	userCache, err := NewTypedBigCache[User](1000)
	if err != nil {
		fmt.Printf("创建用户缓存失败: %v\n", err)
		return
	}
	defer userCache.Close()

	// 使用字符串缓存
	stringCache.Set("greeting", "Hello, World!")
	if value, exists := stringCache.Get("greeting"); exists {
		fmt.Printf("问候语: %s\n", value)
	}

	// 使用用户缓存
	user := User{ID: 1, Name: "李四", Age: 30}
	userCache.Set("user:1", user)
	if cachedUser, exists := userCache.Get("user:1"); exists {
		fmt.Printf("用户: ID=%d, 姓名=%s, 年龄=%d\n",
			cachedUser.ID, cachedUser.Name, cachedUser.Age)
	}
}

// ExampleManager 缓存管理器示例
func ExampleManager() {
	// 创建缓存管理器
	manager := NewManager()

	// 创建不同类型的缓存
	userCache, err := manager.GetCache("users", 1000)
	if err != nil {
		fmt.Printf("创建用户缓存失败: %v\n", err)
		return
	}

	productCache, err := manager.GetCache("products", 500)
	if err != nil {
		fmt.Printf("创建产品缓存失败: %v\n", err)
		return
	}

	sessionCache, err := manager.GetCache("sessions", 200)
	if err != nil {
		fmt.Printf("创建会话缓存失败: %v\n", err)
		return
	}

	// 使用不同的缓存
	userCache.Set("user:1", "用户数据1")
	productCache.Set("product:1", "产品数据1")
	sessionCache.Set("session:1", "会话数据1")

	// 获取所有缓存统计信息
	allStats := manager.GetAllStats()
	for name, stats := range allStats {
		fmt.Printf("缓存 %s: 大小=%d, 命中率=%.2f%%\n",
			name, stats.Hits+stats.Misses, stats.HitRate()*100)
	}

	// 列出所有缓存
	caches := manager.ListCaches()
	fmt.Printf("所有缓存: %v\n", caches)

	// 清理所有缓存
	manager.ClearAll()
}

// ExampleConcurrentUsage 并发使用示例
func ExampleConcurrentUsage() {
	cache, err := NewDefaultBigCache()
	if err != nil {
		fmt.Printf("创建缓存失败: %v\n", err)
		return
	}
	defer cache.Close()

	// 模拟多个goroutine并发访问缓存
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			key := fmt.Sprintf("key:%d", id)
			value := fmt.Sprintf("value:%d", id)

			// 设置缓存
			cache.Set(key, value)

			// 获取缓存
			if v, exists := cache.Get(key); exists {
				fmt.Printf("Goroutine %d: 获取到 %s = %v\n", id, key, v)
			}

			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	fmt.Printf("最终缓存大小: %d\n", cache.Size())
}

// ExampleTTLUsage TTL使用示例
func ExampleTTLUsage() {
	cache, err := NewDefaultBigCache()
	if err != nil {
		fmt.Printf("创建缓存失败: %v\n", err)
		return
	}
	defer cache.Close()

	// 设置不同TTL的缓存
	cache.SetWithTTL("short", "短期数据", 1*time.Second)
	cache.SetWithTTL("medium", "中期数据", 5*time.Second)
	cache.Set("permanent", "永久数据") // 无TTL

	fmt.Println("设置缓存完成，等待过期...")

	// 立即检查
	if value, exists := cache.Get("short"); exists {
		fmt.Printf("短期数据存在: %v\n", value)
	}

	// 等待短期数据过期
	time.Sleep(2 * time.Second)
	if _, exists := cache.Get("short"); !exists {
		fmt.Println("短期数据已过期")
	}

	// 中期数据应该还存在
	if value, exists := cache.Get("medium"); exists {
		fmt.Printf("中期数据存在: %v\n", value)
	}

	// 永久数据应该一直存在
	if value, exists := cache.Get("permanent"); exists {
		fmt.Printf("永久数据存在: %v\n", value)
	}
}

// ExamplePerformance 性能示例
func ExamplePerformance() {
	cache, err := NewDefaultBigCache()
	if err != nil {
		fmt.Printf("创建缓存失败: %v\n", err)
		return
	}
	defer cache.Close()

	// 预热缓存
	fmt.Println("预热缓存...")
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := fmt.Sprintf("value:%d", i)
		cache.Set(key, value)
	}

	// 测试读取性能
	fmt.Println("测试读取性能...")
	start := time.Now()
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("key:%d", i%10000)
		cache.Get(key)
	}
	duration := time.Since(start)

	stats := cache.Stats()
	fmt.Printf("读取100000次耗时: %v\n", duration)
	fmt.Printf("平均每次读取: %v\n", duration/100000)
	fmt.Printf("命中率: %.2f%%\n", stats.HitRate()*100)
}

// ExampleCustomConfig 自定义配置示例
func ExampleCustomConfig() {
	// 创建自定义配置的bigcache
	config := bigcache.Config{
		Shards:             2048,             // 分片数量
		LifeWindow:         30 * time.Minute, // 生命周期
		CleanWindow:        10 * time.Minute, // 清理窗口
		MaxEntriesInWindow: 1000000,          // 最大条目数
		MaxEntrySize:       1000,             // 最大条目大小
		Verbose:            false,            // 不输出详细日志
		HardMaxCacheSize:   0,                // 无硬限制
		Logger:             nil,              // 无日志记录器
	}

	cache, err := NewBigCache(config)
	if err != nil {
		fmt.Printf("创建自定义缓存失败: %v\n", err)
		return
	}
	defer cache.Close()

	// 使用自定义配置的缓存
	cache.Set("custom", "自定义配置缓存")
	if value, exists := cache.Get("custom"); exists {
		fmt.Printf("自定义缓存: %v\n", value)
	}

	fmt.Printf("缓存容量: %d\n", cache.Capacity())
	fmt.Printf("缓存大小: %d\n", cache.Size())
}
