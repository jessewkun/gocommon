# MySQL 数据库模块

本模块提供了 MySQL 数据库的封装，支持连接管理、事务处理、健康检查等功能。

## 功能特性

-   ✅ 支持多实例连接管理
-   ✅ 支持连接池配置
-   ✅ 支持读写分离配置
-   ✅ 支持事务处理
-   ✅ 支持健康检查
-   ✅ 支持优雅关闭
-   ✅ 支持日志记录
-   ✅ 支持基础模型（BaseModel）
-   ✅ 支持自定义时间类型（DateTime）

## 数据模型

### BaseModel 基础模型

提供了基础的数据模型结构，包含通用字段：

```go
type BaseModel struct {
    ID         uint     `gorm:"primarykey" json:"id"`
    CreatedAt  DateTime `gorm:"type:datetime" json:"created_at"`
    ModifiedAt DateTime `gorm:"type:datetime" json:"modified_at"`
}
```

**字段说明：**

-   `ID`: 主键字段，自动递增
-   `CreatedAt`: 创建时间，自动设置
-   `ModifiedAt`: 修改时间，自动更新

**使用示例：**

```go
type User struct {
    BaseModel
    Name  string `gorm:"type:varchar(100)" json:"name"`
    Email string `gorm:"type:varchar(100)" json:"email"`
}

// 创建用户时，CreatedAt 和 ModifiedAt 会自动设置
user := User{
    Name:  "张三",
    Email: "zhangsan@example.com",
}
db.Create(&user)
```

### DateTime 自定义时间类型

提供了自定义的时间类型，支持 JSON 序列化和数据库存储：

```go
type DateTime time.Time
```

**特性：**

-   自动格式化为 "2006-01-02 15:04:05" 格式
-   支持 JSON 序列化和反序列化
-   支持数据库扫描和值转换
-   支持字符串表示

**使用示例：**

```go
type Event struct {
    ID        uint     `gorm:"primarykey" json:"id"`
    Title     string   `json:"title"`
    StartTime DateTime `gorm:"type:datetime" json:"start_time"`
    EndTime   DateTime `gorm:"type:datetime" json:"end_time"`
}

// 创建事件
event := Event{
    Title:     "会议",
    StartTime: DateTime(time.Now()),
    EndTime:   DateTime(time.Now().Add(time.Hour)),
}

// JSON 序列化会自动格式化为字符串
jsonData, _ := json.Marshal(event)
// 输出: {"id":0,"title":"会议","start_time":"2025-06-25 13:45:30","end_time":"2025-06-25 14:45:30"}
```

## 配置说明

### Config 配置结构

```go
type Config struct {
    Dsn                       []string // 数据源
    MaxConn                   int      // 最大连接数
    MaxIdleConn               int      // 最大空闲连接数
    ConnMaxLife               int      // 连接最长持续时间，默认1小时，单位秒
    SlowThreshold             int      // 慢查询阈值，单位毫秒，默认500毫秒
    IgnoreRecordNotFoundError bool     // 是否忽略记录未找到错误
    IsLog                     bool     // 是否记录日志，日志级别为info
}
```

### 配置示例

```go
mysqlConfig := map[string]*Config{
    "default": {
        Dsn:     []string{"user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"},
        MaxConn: 100,
        MaxIdleConn: 25,
        ConnMaxLife: 3600,
        SlowThreshold: 500,
        IgnoreRecordNotFoundError: true,
        IsLog: true,
    },
    "slave": {
        Dsn: []string{
            "user:password@tcp(master:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
            "user:password@tcp(slave1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
            "user:password@tcp(slave2:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        },
        MaxConn: 50,
        MaxIdleConn: 10,
        ConnMaxLife: 3600,
        SlowThreshold: 1000,
        IgnoreRecordNotFoundError: true,
        IsLog: true,
    },
}
```

## 基本使用

### 1. 初始化连接

```go
import "github.com/jessewkun/gocommon/db/mysql"

// 初始化 MySQL 连接
if err := mysql.InitMysql(mysqlConfig); err != nil {
    log.Fatalf("Failed to initialize MySQL: %v", err)
}
```

### 2. 获取数据库连接

```go
// 获取数据库连接
db, err := mysql.GetConn("default")
if err != nil {
    log.Fatalf("Failed to get MySQL connection: %v", err)
}

// 执行查询
var users []User
if err := db.Find(&users).Error; err != nil {
    log.Printf("Failed to query users: %v", err)
}
```

### 3. 事务处理

```go
// 创建事务
tx := mysql.NewTransaction(db)

// 在事务中执行操作
if err := tx.tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.tx.Update(&order).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交事务
if err := tx.Commit(); err != nil {
    return err
}
```

### 4. 健康检查

```go
// 健康检查
healthStatus := mysql.HealthCheck()
for dbName, status := range healthStatus {
    fmt.Printf("MySQL %s health status: %+v\n", dbName, status)
}
```

### 5. 关闭连接

```go
// 关闭连接
if err := mysql.CloseMysql(); err != nil {
    log.Printf("Failed to close MySQL connections: %v", err)
}
```

## 高级功能

### 读写分离

MySQL 模块支持读写分离配置：

```go
mysqlConfig := map[string]*Config{
    "default": {
        Dsn: []string{
            "user:password@tcp(master:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local", // 主库
            "user:password@tcp(slave1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",  // 从库1
            "user:password@tcp(slave2:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",  // 从库2
        },
        MaxConn: 100,
        MaxIdleConn: 25,
        ConnMaxLife: 3600,
        SlowThreshold: 500,
        IgnoreRecordNotFoundError: true,
        IsLog: true,
    },
}
```

-   第一个 DSN 作为主库，用于写操作
-   后续的 DSN 作为从库，用于读操作
-   自动实现读写分离

### 连接池管理

模块自动管理连接池，支持以下配置：

-   `MaxConn`: 最大连接数
-   `MaxIdleConn`: 最大空闲连接数
-   `ConnMaxLife`: 连接最长持续时间

### 慢查询监控

支持慢查询监控和日志记录：

```go
mysqlConfig := map[string]*Config{
    "default": {
        Dsn:           []string{"user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"},
        SlowThreshold: 1000, // 1秒以上的查询会被记录为慢查询
        IsLog:         true, // 启用日志记录
    },
}
```

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
4. **事务使用**: 合理使用事务，确保数据一致性
5. **索引创建**: 建议在查询字段上创建索引以提高性能

## 示例代码

完整的使用示例请参考本目录下的实现文件。

## 依赖

-   `gorm.io/gorm`: GORM ORM 框架
-   `gorm.io/driver/mysql`: MySQL 驱动
-   `gorm.io/plugin/dbresolver`: 数据库解析器插件
