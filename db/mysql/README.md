# MySQL 数据库模块

本模块提供了 MySQL 数据库的封装，支持连接管理、健康检查等功能。

## 功能特性

-   ✅ 支持多实例连接管理
-   ✅ 支持连接池配置（最大连接数、最大空闲连接数、连接最大生命周期、连接最大空闲时间）
-   ✅ 支持读写分离配置
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
    ID         int     `gorm:"primarykey" json:"id"`
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
    Dsn                       []string `mapstructure:"dsn" json:"dsn"`                                                     // 数据源（第一个为主库，后续为从库）
    MaxConn                   int      `mapstructure:"max_conn" json:"max_conn"`                                           // 最大连接数
    MaxIdleConn               int      `mapstructure:"max_idle_conn" json:"max_idle_conn"`                                 // 最大空闲连接数
    ConnMaxLifeTime           int      `mapstructure:"conn_max_life_time" json:"conn_max_life_time"`                       // 连接最长持续时间， 默认1小时，单位秒
    ConnMaxIdleTime           int      `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time"`                       // 连接最大空闲时间， 默认10分钟，单位秒
    SlowThreshold             int      `mapstructure:"slow_threshold" json:"slow_threshold"`                               // 慢查询阈值，单位毫秒，默认500毫秒
    IgnoreRecordNotFoundError bool     `mapstructure:"ignore_record_not_found_error" json:"ignore_record_not_found_error"` // 是否忽略记录未找到错误
    LogLevel                  string   `mapstructure:"log_level" json:"log_level"`                                         // 日志级别：silent/error/warn/info，默认silent
}
```

### 日志级别说明

-   `silent`: 不输出任何 SQL 日志（默认）
-   `error`: 只输出错误日志
-   `warn`: 输出警告和错误日志（包括慢查询）
-   `info`: 输出所有 SQL 日志

### 配置示例

```go
mysqlConfig := map[string]*Config{
    "default": {
        Dsn:     []string{"user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"},
        MaxConn: 100,
        MaxIdleConn: 25,
        ConnMaxLifeTime: 3600,
        ConnMaxIdleTime: 600,
        SlowThreshold: 500,
        IgnoreRecordNotFoundError: true,
        LogLevel: "info", // 输出所有SQL日志
    },
    "slave": {
        Dsn: []string{
            "user:password@tcp(master:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
            "user:password@tcp(slave1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
            "user:password@tcp(slave2:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        },
        MaxConn: 50,
        MaxIdleConn: 10,
        ConnMaxLifeTime: 3600,
        ConnMaxIdleTime: 600,
        SlowThreshold: 1000,
        IgnoreRecordNotFoundError: true,
        LogLevel: "warn", // 只输出慢查询和错误日志
    },
}
```

### 不同环境的配置建议

**开发环境**：

```json
{
    "mysql": {
        "default": {
            "dsn": ["user:pass@tcp(localhost:3306)/dbname"],
            "log_level": "info", // 输出所有SQL，便于调试
            "slow_threshold": 500
        }
    }
}
```

**测试环境**：

```json
{
    "mysql": {
        "default": {
            "dsn": ["user:pass@tcp(localhost:3306)/dbname"],
            "log_level": "warn", // 只关注慢查询和错误
            "slow_threshold": 500
        }
    }
}
```

**生产环境**：

```json
{
    "mysql": {
        "default": {
            "dsn": ["user:pass@tcp(localhost:3306)/dbname"],
            "log_level": "error", // 只关注错误，减少日志量
            "slow_threshold": 500
        }
    }
}
```

## 基本使用

### 1. 初始化连接

```go
import "github.com/jessewkun/gocommon/db/mysql"

// 初始化 MySQL 连接
if err := mysql.Init(); err != nil {
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

### 3. 健康检查

```go
// 健康检查
healthStatus := mysql.HealthCheck()
for dbName, status := range healthStatus {
    fmt.Printf("MySQL %s health status: %+v\n", dbName, status)
}
```

HealthStatus 字段说明：

-   `status`: success/error
-   `error`: 错误信息（当 status=error 时）
-   `latency`: Ping 延迟，毫秒
-   `timestamp`: 检查时间戳（毫秒）
-   `max_open`: 最大连接数
-   `open`: 当前打开连接数
-   `in_use`: 正在使用连接数
-   `idle`: 空闲连接数
-   `wait_count`: 等待连接总次数
-   `wait_time`: 等待连接总时长（纳秒）

### 4. 关闭连接

```go
// 关闭所有连接
if err := mysql.Close(); err != nil {
    log.Printf("Failed to close MySQL connections: %v", err)
} else {
    log.Println("MySQL connections closed successfully")
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
        ConnMaxLifeTime: 3600,
        SlowThreshold: 500,
        IgnoreRecordNotFoundError: true,
        LogLevel: "warn", // 只输出慢查询和错误日志
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
-   `ConnMaxLifeTime`: 连接最长生命周期（秒）
-   `ConnMaxIdleTime`: 连接最大空闲时间（秒）

### 慢查询监控

支持慢查询监控和日志记录：

```go
mysqlConfig := map[string]*Config{
    "default": {
        Dsn:           []string{"user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"},
        SlowThreshold: 1000, // 1秒以上的查询会被记录为慢查询
        LogLevel:      "warn", // 启用慢查询日志记录
    },
}
```

## 错误处理

模块提供了完善的错误处理机制：

1. **连接错误**: 自动重试和日志记录
2. **超时错误**: 可配置的超时时间
3. **健康检查**: 定期检查连接状态

## 性能优化

1. **连接池复用**: 自动管理连接池，避免频繁创建和销毁连接
2. **读写分离**: 支持配置读写分离，提高读取性能
3. **批量操作**: 支持批量插入、更新、删除操作
4. **索引优化**: 建议在查询字段上创建适当的索引

## 注意事项

1. **连接字符串**: 确保连接字符串格式正确，支持认证和 SSL
2. **超时配置**: 根据网络环境调整超时时间
3. **连接池大小**: 根据并发量和服务器资源调整连接池大小
4. **索引创建**: 建议在查询字段上创建索引以提高性能
5. **日志级别**: 根据环境需要合理配置日志级别，避免生产环境输出过多日志

## 示例代码

完整的使用示例请参考本目录下的实现文件。

## 依赖

-   `gorm.io/gorm`: GORM ORM 框架
-   `gorm.io/driver/mysql`: MySQL 驱动
-   `gorm.io/plugin/dbresolver`: 数据库解析器插件
