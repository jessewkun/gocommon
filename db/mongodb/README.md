# MongoDB 数据库模块

本模块提供了 MongoDB 数据库的封装，支持连接管理、事务处理、健康检查等功能。

## 功能特性

-   ✅ 支持多实例连接管理
-   ✅ 支持连接池配置
-   ✅ 支持读写分离配置
-   ✅ 支持事务处理
-   ✅ 支持健康检查
-   ✅ 支持优雅关闭
-   ✅ 支持日志记录

## 配置说明

### Config 配置结构

```go
// 实际结构体名为 Config
// 见 type.go
```

### 配置示例

```go
mongodb.Cfgs = map[string]*mongodb.Config{
    "default": {
        Uris:                   []string{"mongodb://localhost:27017"},
        MaxPoolSize:            100,
        MinPoolSize:            5,
        MaxConnIdleTime:        300,
        ConnectTimeout:         10,
        ServerSelectionTimeout: 5,
        SocketTimeout:          30,
        ReadPreference:         "primary",
        WriteConcern:           "majority",
        IsLog:                  true,
    },
    "replica": {
        Uris:                   []string{"mongodb://localhost:27017,localhost:27018,localhost:27019"},
        MaxPoolSize:            50,
        MinPoolSize:            3,
        MaxConnIdleTime:        300,
        ConnectTimeout:         10,
        ServerSelectionTimeout: 5,
        SocketTimeout:          30,
        ReadPreference:         "secondaryPreferred",
        WriteConcern:           "majority",
        IsLog:                  true,
    },
}
```

> 说明：`Uris` 建议提供**单个完整连接串**（包含用户名/密码、数据库名、参数）。
> 若提供多个条目，模块会将多个 host 合并为一个连接串，并保留**第一个 URI** 的 path/query 参数，其余条目仅作为 host 使用。

## 基本使用

### 1. 初始化连接

```go
import "github.com/jessewkun/gocommon/db/mongodb"

// 先设置全局配置
mongodb.Cfgs = ... // 见上方示例
// 初始化 MongoDB 连接
if err := mongodb.InitMongoDB(); err != nil {
    log.Fatalf("Failed to initialize MongoDB: %v", err)
}
```

### 2. 获取客户端、数据库和集合

```go
// 推荐直接获取集合
collection, err := mongodb.GetMongoCollection("default", "testdb", "users")
if err != nil {
    log.Fatalf("Failed to get collection: %v", err)
}

// 也可分步获取
client, err := mongodb.GetMongoClient("default")
db, err := mongodb.GetMongoDatabase("default", "testdb")
collection := db.Collection("users")
```

### 3. 基本 CRUD 操作

```go
// 插入文档
user := User{
    Name:    "张三",
    Email:   "zhangsan@example.com",
    Age:     25,
    Created: time.Now(),
    Updated: time.Now(),
}
insertResult, err := collection.InsertOne(context.Background(), user)
if err != nil {
    log.Printf("Failed to insert document: %v", err)
}

// 查询文档
var foundUser User
err = collection.FindOne(context.Background(), bson.M{"name": "张三"}).Decode(&foundUser)
if err != nil {
    if err == mongo.ErrNoDocuments {
        fmt.Println("No document found")
    } else {
        log.Printf("Failed to find document: %v", err)
    }
}

// 更新文档
update := bson.M{
    "$set": bson.M{
        "age":    26,
        "updated": time.Now(),
    },
}
updateResult, err := collection.UpdateOne(
    context.Background(),
    bson.M{"name": "张三"},
    update,
)

// 删除文档
deleteResult, err := collection.DeleteOne(context.Background(), bson.M{"name": "张三"})
```

### 4. 事务处理

```go
// 使用事务
client, err := mongodb.GetMongoClient("default")
if err != nil {
    log.Fatalf("Failed to get client: %v", err)
}

err = mongodb.WithTransaction(client, func(sessCtx mongo.SessionContext) error {
    // 在事务中执行操作
    _, err := collection.InsertOne(sessCtx, User{
        Name:    "李四",
        Email:   "lisi@example.com",
        Age:     30,
        Created: time.Now(),
        Updated: time.Now(),
    })
    if err != nil {
        return err
    }
    _, err = collection.InsertOne(sessCtx, User{
        Name:    "王五",
        Email:   "wangwu@example.com",
        Age:     28,
        Created: time.Now(),
        Updated: time.Now(),
    })
    return err
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
} else {
    fmt.Println("Transaction completed successfully")
}
// 注意：单机 MongoDB 不支持事务，需副本集或分片集群
```

### 5. 健康检查

```go
// 健康检查
healthStatus := mongodb.HealthCheck()
for dbName, status := range healthStatus {
    fmt.Printf("MongoDB %s health status: %+v\n", dbName, status)
}
```

### 6. 关闭连接

```go
// 关闭连接
if err := mongodb.CloseMongoDB(); err != nil {
    log.Printf("Failed to close MongoDB connections: %v", err)
}
```

## 高级功能

### 连接池管理

模块自动管理连接池，支持以下配置：

-   `MaxPoolSize`: 最大连接池大小
-   `MinPoolSize`: 最小连接池大小
-   `MaxConnIdleTime`: 连接最大空闲时间

### 读写分离

支持配置读取偏好：

-   `primary`: 只从主节点读取
-   `primaryPreferred`: 优先从主节点读取，主节点不可用时从从节点读取
-   `secondary`: 只从从节点读取
-   `secondaryPreferred`: 优先从从节点读取，从节点不可用时从主节点读取
-   `nearest`: 从最近的节点读取

### 写入关注

支持配置写入关注级别：

-   `majority`: 等待大多数节点确认
-   `1`: 等待一个节点确认
-   `0`: 不等待确认

## 错误处理

模块提供了完善的错误处理机制：

1. **连接错误**: 自动重试和日志记录
2. **事务错误**: 自动回滚
3. **超时错误**: 可配置的超时时间
4. **健康检查**: 定期检查连接状态

## 性能优化

1. **连接池复用**: 自动管理连接池，避免频繁创建和销毁连接
2. **读写分离**: 支持配置读写分离，提高读取性能
3. **批量操作**: 支持批量插入、更新、删除操作
4. **索引优化**: 建议在查询字段上创建适当的索引

## 注意事项

1. **连接字符串**: 确保连接字符串格式正确，支持认证和 SSL
2. **超时配置**: 根据网络环境调整超时时间
3. **连接池大小**: 根据并发量和服务器资源调整连接池大小
4. **事务使用**: MongoDB 事务需要副本集或分片集群
5. **索引创建**: 建议在查询字段上创建索引以提高性能

## 示例代码

完整的使用示例请参考 `example.go` 文件。

## 测试用例

完整的测试用例请参考 `mongodb_test.go` 文件，已兼容单机和副本集环境。

## 依赖

-   `go.mongodb.org/mongo-driver/mongo`: MongoDB Go 驱动
