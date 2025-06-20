# Redis 数据库模块

本模块提供了 Redis 数据库的封装，支持连接管理、健康检查、日志记录等功能。

## 功能特性

-   ✅ 支持多实例连接管理
-   ✅ 支持连接池配置
-   ✅ 支持集群模式
-   ✅ 支持健康检查
-   ✅ 支持日志记录
-   ✅ 支持慢查询监控
-   ✅ 支持发布订阅
-   ✅ 支持事务操作

## 配置说明

### Config 配置结构

```go
type Config struct {
    Addrs              []string // Redis 地址列表 ip:port
    Password           string   // Redis 密码
    Db                 int      // Redis 数据库编号
    IsLog              bool     // 是否记录日志
    PoolSize           int      // 连接池大小，默认500
    IdleTimeout        int      // 空闲连接超时时间，单位秒，默认1秒
    IdleCheckFrequency int      // 空闲连接检查频率，单位秒，默认10秒
    MinIdleConns       int      // 最小空闲连接数，默认3
    MaxRetries         int      // 最大重试次数，默认3
    DialTimeout        int      // 连接超时时间，单位秒，默认2秒
    SlowThreshold      int      // 慢查询阈值，单位毫秒，默认200毫秒
}
```

### 配置示例

```go
redis.Cfgs = map[string]*redis.Config{
    "default": {
        Addrs:              []string{"localhost:6379"},
        Password:           "",
        Db:                 0,
        IsLog:              true,
        PoolSize:           100,
        IdleTimeout:        300,
        IdleCheckFrequency: 60,
        MinIdleConns:       10,
        MaxRetries:         3,
        DialTimeout:        5,
        SlowThreshold:      100,
    },
    "cluster": {
        Addrs:              []string{"localhost:7000", "localhost:7001", "localhost:7002"},
        Password:           "",
        Db:                 0,
        IsLog:              true,
        PoolSize:           50,
        IdleTimeout:        300,
        IdleCheckFrequency: 60,
        MinIdleConns:       5,
        MaxRetries:         3,
        DialTimeout:        5,
        SlowThreshold:      100,
    },
}
```

## 基本使用

### 1. 初始化连接

```go
import "github.com/jessewkun/gocommon/db/redis"

// 先设置全局配置
redis.Cfgs = ... // 见上方示例
// 初始化 Redis 连接
if err := redis.InitRedis(); err != nil {
    log.Fatalf("Failed to initialize Redis: %v", err)
}
```

### 2. 获取 Redis 连接

```go
// 获取 Redis 连接
client, err := redis.GetConn("default")
if err != nil {
    log.Fatalf("Failed to get Redis connection: %v", err)
}

ctx := context.Background()
```

### 3. 基本操作

```go
// 字符串操作
if err := client.Set(ctx, "key1", "value1", time.Hour).Err(); err != nil {
    log.Printf("Failed to set key: %v", err)
}

val, err := client.Get(ctx, "key1").Result()
if err != nil {
    log.Printf("Failed to get key: %v", err)
} else {
    fmt.Printf("Value: %s\n", val)
}

// 哈希表操作
if err := client.HSet(ctx, "user:1", map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
}).Err(); err != nil {
    log.Printf("Failed to set hash: %v", err)
}

userData, err := client.HGetAll(ctx, "user:1").Result()
if err != nil {
    log.Printf("Failed to get hash: %v", err)
} else {
    fmt.Printf("User data: %+v\n", userData)
}
```

### 4. 列表操作

```go
// 从左侧推入元素
if err := client.LPush(ctx, "list1", "item1", "item2", "item3").Err(); err != nil {
    log.Printf("Failed to push to list: %v", err)
}

// 从右侧弹出元素
item, err := client.RPop(ctx, "list1").Result()
if err != nil {
    log.Printf("Failed to pop from list: %v", err)
} else {
    fmt.Printf("Popped item: %s\n", item)
}

// 获取列表长度
length, err := client.LLen(ctx, "list1").Result()
if err != nil {
    log.Printf("Failed to get list length: %v", err)
} else {
    fmt.Printf("List length: %d\n", length)
}
```

### 5. 集合操作

```go
// 添加集合元素
if err := client.SAdd(ctx, "set1", "member1", "member2", "member3").Err(); err != nil {
    log.Printf("Failed to add set members: %v", err)
}

// 检查成员是否存在
exists, err := client.SIsMember(ctx, "set1", "member1").Result()
if err != nil {
    log.Printf("Failed to check set member: %v", err)
} else {
    fmt.Printf("member1 exists: %t\n", exists)
}

// 获取集合所有成员
members, err := client.SMembers(ctx, "set1").Result()
if err != nil {
    log.Printf("Failed to get set members: %v", err)
} else {
    fmt.Printf("Set members: %+v\n", members)
}
```

### 6. 有序集合操作

```go
// 添加有序集合元素
if err := client.ZAdd(ctx, "zset1", &redis.Z{Score: 1.0, Member: "member1"},
    &redis.Z{Score: 2.0, Member: "member2"},
    &redis.Z{Score: 3.0, Member: "member3"}).Err(); err != nil {
    log.Printf("Failed to add zset members: %v", err)
}

// 获取有序集合成员（按分数升序）
zmembers, err := client.ZRange(ctx, "zset1", 0, -1).Result()
if err != nil {
    log.Printf("Failed to get zset members: %v", err)
} else {
    fmt.Printf("ZSet members: %+v\n", zmembers)
}
```

### 7. 管道操作（批量执行）

```go
// 使用管道批量执行命令
pipe := client.Pipeline()
pipe.Set(ctx, "pipe_key1", "value1", time.Hour)
pipe.Set(ctx, "pipe_key2", "value2", time.Hour)
pipe.Get(ctx, "pipe_key1")
pipe.Get(ctx, "pipe_key2")

cmds, err := pipe.Exec(ctx)
if err != nil {
    log.Printf("Failed to execute pipeline: %v", err)
} else {
    fmt.Printf("Pipeline executed successfully, %d commands\n", len(cmds))
}
```

### 8. 事务操作

```go
// 使用事务
// 注意：Redis 事务为乐观锁，不能保证强一致性，适合简单原子操作。
txf := func(tx *redis.Tx) error {
    // 获取当前值
    val, err := tx.Get(ctx, "counter").Result()
    if err != nil && err != redis.Nil {
        return err
    }

    // 在事务中执行操作
    _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        if val == "" {
            pipe.Set(ctx, "counter", 1, time.Hour)
        } else {
            pipe.Incr(ctx, "counter")
        }
        return nil
    })
    return err
}

// 执行事务
err := client.Watch(ctx, txf, "counter")
if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

### 9. 发布订阅

```go
// 创建订阅
pubsub := client.Subscribe(ctx, "channel1")
defer pubsub.Close()

// 发布消息
if err := client.Publish(ctx, "channel1", "Hello Redis!").Err(); err != nil {
    log.Printf("Failed to publish message: %v", err)
}

// 接收消息（非阻塞）
msg, err := pubsub.ReceiveTimeout(ctx, time.Second)
if err != nil {
    log.Printf("Failed to receive message: %v", err)
} else {
    fmt.Printf("Received message: %+v\n", msg)
}
```

### 10. 健康检查

```go
// 健康检查
healthStatus := redis.HealthCheck()
for dbName, status := range healthStatus {
    fmt.Printf("Redis %s health status: %+v\n", dbName, status)
}
```

## 高级功能

### 连接池管理

模块自动管理连接池，支持以下配置：

-   `PoolSize`: 连接池大小
-   `IdleTimeout`: 空闲连接超时时间
-   `MinIdleConns`: 最小空闲连接数
-   `IdleCheckFrequency`: 空闲连接检查频率

### 集群模式

支持 Redis 集群模式，可以配置多个节点：

```go
// 注意：集群模式下需配置所有节点地址，且部分命令不支持跨 slot 操作。
redisConfig := map[string]*Config{
    "cluster": {
        Addrs: []string{
            "localhost:7000",
            "localhost:7001",
            "localhost:7002",
        },
        PoolSize: 50,
        // ... 其他配置
    },
}
```

### 慢查询监控

支持慢查询监控和日志记录：

```go
redisConfig := map[string]*Config{
    "default": {
        Addrs:         []string{"localhost:6379"},
        IsLog:         true,
        SlowThreshold: 100, // 100毫秒以上的查询会被记录为慢查询
    },
}
```

## 错误处理

模块提供了完善的错误处理机制：

1. **连接错误**: 自动重试和日志记录
2. **超时错误**: 可配置的超时时间
3. **健康检查**: 定期检查连接状态
4. **慢查询监控**: 记录慢查询日志

## 性能优化

1. **连接池复用**: 自动管理连接池，避免频繁创建和销毁连接
2. **管道操作**: 支持批量执行命令，减少网络往返
3. **事务支持**: 支持 Redis 事务，确保数据一致性
4. **集群模式**: 支持 Redis 集群，提高可用性和性能

## 注意事项

1. **连接字符串**: 确保连接地址格式正确
2. **超时配置**: 根据网络环境调整超时时间
3. **连接池大小**: 根据并发量和服务器资源调整连接池大小
4. **密码安全**: 使用安全的密码，避免明文存储
5. **集群配置**: 集群模式下需要配置所有节点地址

## 示例代码

完整的使用示例请参考 example.go 文件。

## 测试用例

完整的测试用例请参考 redis_test.go 文件。

## 依赖

-   `github.com/go-redis/redis/v8`: Redis Go 驱动
