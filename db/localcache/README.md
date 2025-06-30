# BigCache 本地缓存模块

这是一个基于 [BigCache](https://github.com/allegro/bigcache) 的高性能本地缓存模块，提供了统一的缓存接口、TTL 支持、类型安全缓存和缓存管理等功能。

## 特性

-   **高性能**: 基于 BigCache，零 GC 压力，适合大容量缓存
-   **TTL 支持**: 支持设置缓存项的过期时间
-   **类型安全**: 提供泛型支持的类型安全缓存
-   **并发安全**: 使用读写锁保证并发安全
-   **统计信息**: 提供命中率、淘汰次数等统计信息
-   **缓存管理**: 支持多个命名缓存的管理
-   **自动清理**: 后台自动清理过期项
-   **JSON 序列化**: 支持复杂数据结构的序列化和反序列化

## 三种缓存接口的区别

本模块提供了三种不同的缓存接口，满足不同的使用需求：

### 1. Cache 接口（基础接口）

**Cache** 是基础的缓存接口，使用 `interface{}` 类型：

```go
type Cache interface {
    Get(key string) (interface{}, bool)                    // 返回 interface{}
    Set(key string, value interface{}) error               // 接受 interface{}
    SetWithTTL(key string, value interface{}, ttl time.Duration) error
    Delete(key string) bool
    Clear()
    Size() int
    Capacity() int
    Stats() Stats
    Close() error
}
```

**特点：**

-   使用 `interface{}` 类型，需要类型断言
-   可以存储任意类型的数据
-   类型安全性较差，容易出现运行时错误
-   适合简单的缓存需求

**使用示例：**

```go
cache, _ := NewDefaultBigCache()
cache.Set("user", map[string]interface{}{"id": 1, "name": "张三"})

// 需要类型断言
if value, exists := cache.Get("user"); exists {
    if user, ok := value.(map[string]interface{}); ok {
        fmt.Println(user["name"])
    }
}
```

### 2. TypedCache 接口（类型安全接口）

**TypedCache** 是类型安全的缓存接口，使用 Go 泛型：

```go
type TypedCache[T any] interface {
    Get(key string) (T, bool)                              // 返回具体类型 T
    Set(key string, value T) error                         // 接受具体类型 T
    SetWithTTL(key string, value T, ttl time.Duration) error
    Delete(key string) bool
    Clear()
    Size() int
    Capacity() int
    Stats() Stats
    Close() error
}
```

**特点：**

-   使用泛型 `T`，编译时类型安全
-   不需要类型断言
-   更好的 IDE 支持和代码提示
-   避免运行时类型错误
-   **完全保留 Stats 功能**：与 Cache 接口具有相同的统计功能
-   适合需要类型安全的场景

**使用示例：**

```go
// 字符串类型缓存
stringCache, _ := NewTypedBigCache[string](1000)
stringCache.Set("greeting", "Hello")
if value, exists := stringCache.Get("greeting"); exists {
    fmt.Println(value) // value 是 string 类型
}

// 用户类型缓存
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

userCache, _ := NewTypedBigCache[User](1000)
user := User{ID: 1, Name: "张三"}
userCache.Set("user:1", user)
if cachedUser, exists := userCache.Get("user:1"); exists {
    fmt.Println(cachedUser.Name) // cachedUser 是 User 类型
}

// TypedCache 同样支持统计功能
stats := userCache.Stats()
fmt.Printf("命中率: %.2f%%\n", stats.HitRate()*100)
```

### 3. Manager 管理器（多缓存管理）

**Manager** 是缓存管理器，用于管理多个命名缓存：

```go
type Manager struct {
    caches map[string]Cache
    mutex  sync.RWMutex
}
```

**特点：**

-   管理多个不同的缓存实例
-   每个缓存有独立的名称和配置
-   提供统一的统计和管理接口
-   支持缓存的创建、删除、查询等操作
-   适合需要管理多个缓存的场景

**使用示例：**

```go
manager := NewManager()

// 创建不同类型的缓存
userCache, _ := manager.GetCache("users", 1000)      // 用户缓存
productCache, _ := manager.GetCache("products", 500)  // 产品缓存
sessionCache, _ := manager.GetCache("sessions", 200)  // 会话缓存

// 使用不同的缓存
userCache.Set("user:1", "用户数据")
productCache.Set("product:1", "产品数据")
sessionCache.Set("session:1", "会话数据")

// 获取所有缓存统计
allStats := manager.GetAllStats()
for name, stats := range allStats {
    fmt.Printf("缓存 %s: 命中率=%.2f%%\n", name, stats.HitRate()*100)
}

// 列出所有缓存
caches := manager.ListCaches()
fmt.Printf("所有缓存: %v\n", caches)
```

### 接口对比表

| 特性           | Cache               | TypedCache        | Manager            |
| -------------- | ------------------- | ----------------- | ------------------ |
| **类型安全**   | ❌ 使用 interface{} | ✅ 编译时类型安全 | ❌ 管理 Cache 接口 |
| **类型断言**   | 需要                | 不需要            | 需要               |
| **泛型支持**   | ❌                  | ✅                | ❌                 |
| **统计功能**   | ✅ 完整支持         | ✅ 完整支持       | ✅ 统一管理        |
| **多缓存管理** | ❌                  | ❌                | ✅                 |
| **命名空间**   | ❌                  | ❌                | ✅                 |
| **使用场景**   | 简单缓存需求        | 类型安全需求      | 多缓存管理需求     |

### 选择建议

1. **简单场景**：直接使用 `Cache` 接口

    ```go
    cache, _ := NewDefaultBigCache()
    ```

2. **类型安全场景**：使用 `TypedCache`

    ```go
    userCache, _ := NewTypedBigCache[User](1000)
    ```

3. **多缓存管理场景**：使用 `Manager`

    ```go
    manager := NewManager()
    userCache, _ := manager.GetCache("users", 1000)
    ```

4. **组合使用**：Manager + TypedCache
    ```go
    // 虽然 Manager 不直接支持 TypedCache，但可以这样使用
    manager := NewManager()
    cache, _ := manager.GetCache("users", 1000)
    // 然后手动进行类型转换
    ```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/jessewkun/gocommon/db/localcache"
)

func main() {
    // 创建默认配置的bigcache
    cache, err := localcache.NewDefaultBigCache()
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    // 设置缓存
    cache.Set("key1", "value1")

    // 设置带TTL的缓存
    cache.SetWithTTL("key2", "value2", 30*time.Minute)

    // 获取缓存
    if value, exists := cache.Get("key1"); exists {
        fmt.Printf("获取到值: %v\n", value)
    }

    // 删除缓存
    cache.Delete("key1")

    // 获取统计信息
    stats := cache.Stats()
    fmt.Printf("命中率: %.2f%%\n", stats.HitRate()*100)
}
```

### 类型安全缓存

```go
// 创建字符串类型缓存
stringCache, err := localcache.NewTypedBigCache[string](1000)
if err != nil {
    panic(err)
}
defer stringCache.Close()

// 创建用户结构体类型缓存
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

userCache, err := localcache.NewTypedBigCache[User](1000)
if err != nil {
    panic(err)
}
defer userCache.Close()

// 使用类型安全缓存
user := User{ID: 1, Name: "张三", Age: 25}
userCache.Set("user:1", user)

if cachedUser, exists := userCache.Get("user:1"); exists {
    fmt.Printf("用户: %s, 年龄: %d\n", cachedUser.Name, cachedUser.Age)
}

// TypedCache 同样支持完整的统计功能
stats := userCache.Stats()
fmt.Printf("缓存统计: 命中=%d, 未命中=%d, 命中率=%.2f%%\n",
    stats.Hits, stats.Misses, stats.HitRate()*100)
```

### 缓存管理器

```go
// 创建缓存管理器
manager := localcache.NewManager()

// 创建不同类型的缓存
userCache, err := manager.GetCache("users", 1000)
if err != nil {
    panic(err)
}

productCache, err := manager.GetCache("products", 500)
if err != nil {
    panic(err)
}

// 使用缓存
userCache.Set("user:1", "用户数据")
productCache.Set("product:1", "产品数据")

// 获取所有缓存统计信息
allStats := manager.GetAllStats()
for name, stats := range allStats {
    fmt.Printf("缓存 %s: 命中率=%.2f%%\n", name, stats.HitRate()*100)
}

// 列出所有缓存
caches := manager.ListCaches()
fmt.Printf("所有缓存: %v\n", caches)
```

### 自定义配置

```go
// 创建自定义配置的bigcache
config := bigcache.Config{
    Shards:             2048,           // 分片数量
    LifeWindow:         30 * time.Minute, // 生命周期
    CleanWindow:        10 * time.Minute, // 清理窗口
    MaxEntriesInWindow: 1000000,        // 最大条目数
    MaxEntrySize:       1000,           // 最大条目大小
    Verbose:            false,          // 不输出详细日志
    HardMaxCacheSize:   0,              // 无硬限制
    Logger:             nil,            // 无日志记录器
}

cache, err := localcache.NewBigCache(config)
if err != nil {
    panic(err)
}
defer cache.Close()
```

## API 参考

### Cache 接口

```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}) error
    SetWithTTL(key string, value interface{}, ttl time.Duration) error
    Delete(key string) bool
    Clear()
    Size() int
    Capacity() int
    Stats() Stats
    Close() error
}
```

### TypedCache 接口

```go
type TypedCache[T any] interface {
    Get(key string) (T, bool)
    Set(key string, value T) error
    SetWithTTL(key string, value T, ttl time.Duration) error
    Delete(key string) bool
    Clear()
    Size() int
    Capacity() int
    Stats() Stats
    Close() error
}
```

### Stats 结构体

```go
type Stats struct {
    Hits        int64 `json:"hits"`        // 命中次数
    Misses      int64 `json:"misses"`      // 未命中次数
    Evictions   int64 `json:"evictions"`   // 淘汰次数
    Expirations int64 `json:"expirations"` // 过期次数
}

func (s Stats) HitRate() float64 // 计算命中率
```

## 性能特性

-   **零 GC 压力**: BigCache 使用内存映射，避免 GC 压力
-   **高并发**: 支持高并发读写操作
-   **内存效率**: 自动管理内存使用，避免内存泄漏
-   **快速访问**: O(1) 时间复杂度的基本操作
-   **大容量**: 支持百万级别的缓存条目

## BigCache 优势

1. **零 GC**: 使用内存映射，不会产生 GC 压力
2. **高性能**: 专门为高并发场景优化
3. **内存友好**: 自动管理内存，避免内存泄漏
4. **大容量**: 支持存储大量数据
5. **分片设计**: 使用分片减少锁竞争

## 使用建议

1. **合理设置容量**: 根据实际需求设置 `MaxEntriesInWindow`
2. **使用 TTL**: 对于有生命周期限制的数据，建议设置合适的 TTL
3. **监控统计**: 定期检查缓存统计信息，优化缓存策略
4. **类型安全**: 优先使用类型安全缓存，避免类型转换错误
5. **资源管理**: 记得调用 Close() 方法释放资源
6. **配置优化**: 根据实际场景调整 BigCache 的配置参数

## 示例

更多使用示例请参考 `example.go` 文件，包含：

-   基本使用示例
-   类型安全缓存示例
-   缓存管理器示例
-   并发使用示例
-   TTL 使用示例
-   性能测试示例
-   自定义配置示例

## 测试

运行测试：

```bash
go test ./db/localcache -v
```

运行性能测试：

```bash
go test ./db/localcache -bench=.
```

## 依赖

-   `github.com/allegro/bigcache` - 高性能缓存库
-   `github.com/jessewkun/gocommon/common` - 通用错误处理

## 注意事项

1. **删除操作**: BigCache 本身不支持删除操作，我们通过设置空值来模拟删除
2. **清空操作**: BigCache 不支持清空操作，我们只重置统计信息
3. **TTL 实现**: TTL 是通过在数据中嵌入过期时间实现的，不是 BigCache 原生功能
4. **序列化开销**: 复杂数据结构需要 JSON 序列化，会有一定的性能开销
5. **内存使用**: BigCache 会预分配内存，实际内存使用可能超过预期
6. **统计功能**: TypedCache 完全保留了 Cache 接口的所有统计功能，包括命中率、淘汰次数等

## 技术选型

如果您正在考虑选择不同的缓存库，可以参考我们的技术选型文档：

📖 **[技术选型对比](./TECHNOLOGY_CHOICE.md)** - 详细对比 BigCache、Ristretto 和 FreeCache 的优缺点和适用场景

该文档包含：

-   三个缓存库的详细对比表
-   配置复杂度分析
-   性能对比数据
-   选择建议和迁移策略
-   参考资料和论文链接
