# Nacos 模块

Nacos 模块提供了配置管理和服务发现功能，支持多实例管理(因本地 nacos 启动失败，该 package 未经测试)。

## 特性

-   ✅ 支持多实例配置管理（通过显式 `Manager` 或全局便利层）
-   ✅ 支持灵活的初始化方式和资源管理
-   ✅ 线程安全的连接管理
-   ✅ 配置管理和服务发现功能
-   ✅ 可配置的 Nacos 客户端参数（日志目录、缓存目录、日志级别等）
-   ✅ 更健壮的错误处理和资源清理
-   ✅ 更好的可测试性

## 配置说明

### Config 配置结构

```go
type Config struct {
    Host                string `mapstructure:"host" json:"host"`                           // Nacos 服务器地址
    Port                uint64 `mapstructure:"port" json:"port"`                         // Nacos 服务器端口
    Namespace           string `mapstructure:"namespace" json:"namespace"`               // 命名空间
    Group               string `mapstructure:"group" json:"group"`                       // 配置组
    Username            string `mapstructure:"username" json:"username"`                   // 用户名
    Password            string `mapstructure:"password" json:"password"`                   // 密码
    Timeout             int    `mapstructure:"timeout" json:"timeout"`                   // 超时时间(毫秒)
    NotLoadCacheAtStart bool   `mapstructure:"not_load_cache_at_start" json:"not_load_cache_at_start"` // 是否在启动时加载缓存，默认为 true
    LogDir              string `mapstructure:"log_dir" json:"log_dir"`                   // Nacos 客户端日志存储路径，默认为 /tmp/nacos/log
    CacheDir            string `mapstructure:"cache_dir" json:"cache_dir"`               // Nacos 客户端缓存存储路径，默认为 /tmp/nacos/cache
    LogLevel            string `mapstructure:"log_level" json:"log_level"`               // Nacos 客户端日志级别，默认为 warn
}
```

### 配置示例

#### 全局配置方式 (`Cfgs`)

```go
// nacos.Cfgs 通常通过 config 包加载，或者在程序启动时手动设置
nacos.Cfgs = map[string]*nacos.Config{
    "default": {
        Host:      "localhost",
        Port:      8848,
        Namespace: "public",
        Group:     "DEFAULT_GROUP",
        Timeout:   5000,
        LogLevel:  "info", // 可以自定义日志级别
    },
    "dev": {
        Host:      "dev-nacos.example.com",
        Port:      8848,
        Namespace: "dev",
        Group:     "DEFAULT_GROUP",
        Username:  "dev-user",
        Password:  "dev-pass",
        Timeout:   5000,
        LogDir:    "/var/log/nacos/dev", // 可以自定义日志目录
    },
}
```

#### 配置文件方式

创建配置文件 `config.toml`：

```toml
[nacos]
[nacos.default]
host = "localhost"
port = 8848
namespace = "public"
group = "DEFAULT_GROUP"
timeout = 5000
logLevel = "info"

[nacos.dev]
host = "dev-nacos.example.com"
port = 8848
namespace = "dev"
group = "DEFAULT_GROUP"
username = "dev-user"
password = "dev-pass"
timeout = 5000
logDir = "/var/log/nacos/dev"
```

## 快速开始

Nacos 模块提供两种使用方式：**显式管理器模式**（推荐，更灵活、可测试）和**全局便利层**（简化常用场景）。

### 1. 显式管理器使用 (推荐)

此模式允许您创建独立的 `Manager` 实例，适合需要多套 Nacos 连接配置、或者重视可测试性的场景。

```go
package main

import (
	"context"
	"log"
	"github.com/jessewkun/gocommon/nacos"
)

func main() {
	// 定义 Nacos 客户端配置
	configs := map[string]*nacos.Config{
		"default": {
			Host:      "localhost",
			Port:      8848,
			Namespace: "public",
			Group:     "DEFAULT_GROUP",
			Timeout:   5000,
			LogLevel:  "info",
		},
		"dev": {
			Host:      "dev-nacos.example.com",
			Port:      8848,
			Namespace: "dev",
			Group:     "DEFAULT_GROUP",
			Username:  "dev-user",
			Password:  "dev-pass",
			Timeout:   5000,
		},
	}

	// 创建并初始化 Nacos 管理器
	// NewManager 会尝试连接所有配置的实例
	mgr, err := nacos.NewManager(configs)
	if err != nil {
		log.Fatalf("Failed to create Nacos Manager: %v", err)
	}
	// 确保在程序结束时关闭所有连接
	defer func() {
		if closeErr := mgr.Close(); closeErr != nil {
			log.Printf("Error closing Nacos Manager: %v", closeErr)
		}
	}()

	// 获取默认客户端
	defaultClient, err := mgr.GetClient("default")
	if err != nil {
		log.Fatalf("Failed to get default client: %v", err)
	}

	// 获取开发环境客户端
	devClient, err := mgr.GetClient("dev")
	if err != nil {
		log.Fatalf("Failed to get dev client: %v", err)
	}

	// 使用客户端进行配置管理
	err = defaultClient.PublishConfig("app-config.json", `{"env": "development"}`)
	if err != nil {
		log.Printf("Failed to publish config: %v", err)
	}

	content, err := defaultClient.GetConfig("app-config.json")
	if err != nil {
		log.Printf("Failed to get config: %v", err)
	} else {
		log.Printf("Config from default: %s", content)
	}

	// 使用 devClient 进行服务发现 (示例)
	err = devClient.RegisterService("my-service", "127.0.0.1", 8080, nil)
	if err != nil {
		log.Printf("Failed to register service with devClient: %v", err)
	}
}
```

### 2. 全局便利层使用

此模式适用于应用程序中只需要一套 Nacos 连接配置的简单场景。它依赖于 `nacos.Cfgs` 全局变量和 `config` 包的自动加载机制。

```go
package main

import (
    "log"
    "github.com/jessewkun/gocommon/nacos"
    // 通常 config 包会自动初始化 nacos.Cfgs 并调用 nacos.Init()
    // import "github.com/jessewkun/gocommon/config"
)

func main() {
    // 如果没有使用 config 包自动加载，则需要手动设置 Cfgs
    nacos.Cfgs = map[string]*nacos.Config{
        "default": {
            Host:      "localhost",
            Port:      8848,
            Namespace: "public",
            Group:     "DEFAULT_GROUP",
            Timeout:   5000,
            LogLevel:  "info",
        },
    }

    // 初始化所有配置的客户端 (如果 config 包未自动调用)
    if err := nacos.Init(); err != nil {
        log.Fatalf("Failed to initialize Nacos: %v", err)
    }
    // 确保在程序结束时关闭所有连接
    defer func() {
        if closeErr := nacos.Close(); closeErr != nil {
            log.Printf("Error closing global Nacos clients: %v", closeErr)
        }
    }()

    // 获取客户端
    client, err := nacos.GetClient("default") // 注意：现在是 GetClient
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

## API 参考

### 配置结构

请参考上方的 [Config 配置结构](#config-配置结构)

### 主要函数（全局便利层）

-   `nacos.Init() error` - 初始化所有 `nacos.Cfgs` 中定义的客户端，并赋值给全局 `defaultManager`。
-   `nacos.GetClient(name string) (*Client, error)` - 从全局 `defaultManager` 获取指定名称的客户端。
-   `nacos.Close() error` - 关闭全局 `defaultManager` 中所有客户端连接。
-   `nacos.DefaultConfig() *Config` - 获取一个包含默认值的 `Config` 实例。

### 主要函数（显式管理器模式）

-   `nacos.NewManager(configs map[string]*Config) (*Manager, error)` - 创建并初始化一个 `Manager` 实例，根据传入的 `configs` 管理 Nacos 客户端。
-   `(*Manager).GetClient(name string) (*Client, error)` - 从该 `Manager` 实例中获取指定名称的客户端。
-   `(*Manager).Close() error` - 关闭该 `Manager` 实例管理的所有客户端连接。

### 客户端方法 (`*nacos.Client`)

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
-   `GetServiceOne(serviceName string) (*ServiceInfo, error)` - 获取一个健康的服务实例
-   `SubscribeService(serviceName string, onUpdate func([]ServiceInfo)) error` - 订阅服务变化
-   `UnsubscribeService(serviceName string) error` - 取消订阅服务
-   `GetServices(pageNo, pageSize int) (model.ServiceList, error)` - 分页获取所有服务信息（**注意：此方法返回 Nacos SDK 内部 ServiceList 类型**）


## 示例

完整的使用示例请参考 `example.go` 文件（该文件会更新以反映新的使用方式）。

## 测试

运行测试：

```bash
go test -v ./...
```

运行基准测试：

```bash
go test -bench=. ./...
```
新架构显著提升了测试的稳定性和隔离性。
