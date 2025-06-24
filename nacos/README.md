# Nacos 模块

Nacos 模块提供了配置管理和服务发现功能，支持多实例管理。

## 特性

-   支持多实例配置管理
-   自动配置加载和初始化
-   线程安全的连接管理
-   配置管理和服务发现功能
-   支持配置文件自动加载

## 快速开始

### 1. 基本使用

```go
package main

import (
    "log"
    "github.com/jessewkun/gocommon/nacos"
)

func main() {
    // 设置配置
    nacos.Cfgs = map[string]*nacos.Config{
        "default": {
            Host:      "localhost",
            Port:      8848,
            Namespace: "public",
            Group:     "DEFAULT_GROUP",
            Timeout:   5000,
        },
    }

    // 初始化
    if err := nacos.Init(); err != nil {
        log.Fatalf("Failed to initialize Nacos: %v", err)
    }
    defer nacos.Close()

    // 获取客户端
    client, err := nacos.GetConn("default")
    if err != nil {
        log.Fatalf("Failed to get client: %v", err)
    }

    // 发布配置
    err = client.PublishConfig("app-config.json", `{"env": "development"}`)
    if err != nil {
        log.Printf("Failed to publish config: %v", err)
    }

    // 获取配置
    content, err := client.GetConfig("app-config.json")
    if err != nil {
        log.Printf("Failed to get config: %v", err)
    } else {
        log.Printf("Config: %s", content)
    }
}
```

### 2. 多实例使用

```go
// 设置多个环境配置
nacos.Cfgs = map[string]*nacos.Config{
    "dev": {
        Host:      "dev-nacos.example.com",
        Port:      8848,
        Namespace: "dev",
        Group:     "DEFAULT_GROUP",
        Username:  "dev-user",
        Password:  "dev-pass",
        Timeout:   5000,
    },
    "prod": {
        Host:      "prod-nacos.example.com",
        Port:      8848,
        Namespace: "prod",
        Group:     "DEFAULT_GROUP",
        Username:  "prod-user",
        Password:  "prod-pass",
        Timeout:   5000,
    },
}

// 初始化所有实例
if err := nacos.Init(); err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}
defer nacos.Close()

// 使用不同环境的客户端
devClient, _ := nacos.GetConn("dev")
prodClient, _ := nacos.GetConn("prod")

// 在开发环境发布配置
devClient.PublishConfig("dev-config.json", `{"env": "development"}`)

// 在生产环境发布配置
prodClient.PublishConfig("prod-config.json", `{"env": "production"}`)
```

### 3. 配置文件方式

创建配置文件 `config.toml`：

```toml
[nacos]
[nacos.default]
host = "localhost"
port = 8848
namespace = "public"
group = "DEFAULT_GROUP"
timeout = 5000

[nacos.dev]
host = "dev-nacos.example.com"
port = 8848
namespace = "dev"
group = "DEFAULT_GROUP"
username = "dev-user"
password = "dev-pass"
timeout = 5000

[nacos.prod]
host = "prod-nacos.example.com"
port = 8848
namespace = "prod"
group = "DEFAULT_GROUP"
username = "prod-user"
password = "prod-pass"
timeout = 5000
```

```go
// 配置会自动加载，只需要调用初始化
if err := nacos.Init(); err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}
```

## API 参考

### 配置结构

```go
type Config struct {
    Host      string // nacos服务器地址
    Port      uint64 // nacos服务器端口
    Namespace string // 命名空间
    Group     string // 配置组
    Username  string // 用户名
    Password  string // 密码
    Timeout   int    // 超时时间(毫秒)
}
```

### 主要函数

-   `Init() error` - 初始化所有配置的客户端
-   `GetConn(name string) (*Client, error)` - 获取指定名称的客户端
-   `Close() error` - 关闭所有客户端连接
-   `DefaultConfig() *Config` - 获取默认配置

### 客户端方法

#### 配置管理

-   `GetConfig(dataId string) (string, error)` - 获取配置
-   `PublishConfig(dataId, content string) error` - 发布配置
-   `DeleteConfig(dataId string) error` - 删除配置
-   `ListenConfig(dataId string, onChange func(string, string, string)) error` - 监听配置变化
-   `CancelListenConfig(dataId string) error` - 取消监听配置

#### 服务管理

-   `RegisterService(serviceName, ip string, port uint64, metadata map[string]string) error` - 注册服务
-   `DeregisterService(serviceName, ip string, port uint64) error` - 注销服务
-   `GetService(serviceName string) ([]ServiceInfo, error)` - 获取服务实例
-   `GetServiceOne(serviceName string) (*ServiceInfo, error)` - 获取一个服务实例
-   `SubscribeService(serviceName string, onUpdate func([]ServiceInfo)) error` - 订阅服务变化
-   `UnsubscribeService(serviceName string) error` - 取消订阅服务

## 改进说明

相比之前的实现，新版本有以下改进：

1. **统一配置管理**：使用 `Cfgs` map 统一管理所有实例配置
2. **自动初始化**：通过 `config.Register` 和 `config.RegisterCallback` 实现配置自动加载
3. **连接复用**：避免重复创建相同配置的客户端
4. **线程安全**：使用读写锁保证并发安全
5. **错误处理**：提供详细的错误信息和日志记录
6. **资源管理**：统一的连接关闭和资源清理

## 示例

完整的使用示例请参考 `example.go` 文件。

## 测试

运行测试：

```bash
go test -v
```

运行基准测试：

```bash
go test -bench=.
```
