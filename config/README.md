# 配置管理模块

## 功能简介

配置管理模块提供灵活的配置加载和管理解决方案，支持多种配置格式、模块化配置注册、依赖管理、热更新等功能。

- ✅ 支持 TOML、JSON、YAML 格式配置文件
- ✅ 模块化配置注册和自动加载
- ✅ 依赖关系管理（拓扑排序）
- ✅ 配置热更新（无需重启服务）
- ✅ 跨模块依赖注入支持
- ✅ 基础配置管理（运行模式、端口、域名等）
- ✅ 自动初始化回调函数执行

## 依赖

- `github.com/spf13/viper`：配置管理库
- `github.com/fsnotify/fsnotify`：文件监听库（用于热更新）

## 基础配置

### BaseConfig 结构

```go
type BaseConfig struct {
    Mode    string `mapstructure:"mode" json:"mode"`         // 运行模式: debug 开发, release 生产, test 测试
    Port    string `mapstructure:"port" json:"port"`         // 服务端口, 默认 ":8000"
    AppName string `mapstructure:"app_name" json:"app_name"` // 服务标题, 默认 "Service"
    Domain  string `mapstructure:"domain" json:"domain"`     // 服务域名, 默认 "http://localhost:8000"
}
```

### 基础配置示例

**TOML 格式：**

```toml
mode = "debug"
port = ":8000"
app_name = "MyService"
domain = "http://localhost:8000"
```

**JSON 格式：**

```json
{
  "mode": "debug",
  "port": ":8000",
  "app_name": "MyService",
  "domain": "http://localhost:8000"
}
```

**YAML 格式：**

```yaml
mode: debug
port: ":8000"
app_name: MyService
domain: "http://localhost:8000"
```

## 基本使用

### 1. 初始化配置

在应用启动时调用 `Init` 函数加载配置：

```go
import "github.com/jessewkun/gocommon/config"

// 加载配置文件
cfg, err := config.Init("./config.toml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 使用基础配置
fmt.Printf("Mode: %s, Port: %s\n", cfg.Mode, cfg.Port)
```

### 2. 访问基础配置

```go
// 直接访问全局配置实例
fmt.Printf("App Name: %s\n", config.Cfg.AppName)
fmt.Printf("Domain: %s\n", config.Cfg.Domain)
```

## 模块化配置

### 注册模块配置

在模块的 `init()` 函数中注册配置：

```go
package mymodule

import "github.com/jessewkun/gocommon/config"

type Config struct {
    Host     string `mapstructure:"host" json:"host"`
    Port     int    `mapstructure:"port" json:"port"`
    Timeout  int    `mapstructure:"timeout" json:"timeout"`
}

var Cfg = &Config{
    Host:    "localhost",
    Port:    8080,
    Timeout: 5,
}

func init() {
    // 注册模块配置，key 为模块名，cfgPtr 为配置结构体指针
    config.Register("mymodule", Cfg)
}
```

### 配置文件示例

**TOML 格式：**

```toml
[mymodule]
host = "localhost"
port = 8080
timeout = 5
```

**JSON 格式：**

```json
{
  "mymodule": {
    "host": "localhost",
    "port": 8080,
    "timeout": 5
  }
}
```

### 注册初始化回调

如果模块需要在配置加载后执行初始化逻辑，可以注册回调函数：

```go
func init() {
    config.Register("mymodule", Cfg)
    // 注册初始化回调，dependencies 指定依赖的模块
    config.RegisterCallback("mymodule", Init, "config", "http", "log")
}

// Init 初始化函数，会在配置加载后自动调用
func Init() error {
    // 执行初始化逻辑
    client = NewClient(Cfg.Host, Cfg.Port)
    return nil
}
```

### 依赖管理

配置模块支持模块间的依赖关系管理，使用拓扑排序确保依赖模块先初始化：

```go
// 模块 A 依赖 config 和 log
config.RegisterCallback("moduleA", InitA, "config", "log")

// 模块 B 依赖 config、log 和 moduleA
config.RegisterCallback("moduleB", InitB, "config", "log", "moduleA")
```

**特性：**

- 自动检测循环依赖
- 按依赖顺序执行初始化回调
- 如果依赖模块没有回调函数，会被自动忽略（因为配置已加载）

**依赖执行顺序：**

1. 基础配置（`config`）和没有回调的模块配置先加载
2. 按拓扑排序顺序执行有回调的模块初始化
3. 最后执行跨模块依赖注入函数

### 跨模块依赖注入

对于复杂的跨模块依赖，可以使用注入器函数：

```go
// 注册注入器函数，在所有模块回调执行完毕后调用
config.RegisterInjector(func() error {
    // 处理跨模块依赖注入
    logger.SetAlerter(alarm.NewSender())
    return nil
})
```

## 配置热更新

### 实现 HotReloadable 接口

模块配置结构体可以实现 `HotReloadable` 接口以支持热更新：

```go
import (
    "github.com/jessewkun/gocommon/config"
    "github.com/spf13/viper"
)

type Config struct {
    Timeout int `mapstructure:"timeout" json:"timeout"`
}

// Reload 实现 HotReloadable 接口
func (c *Config) Reload(v *viper.Viper) error {
    if err := v.UnmarshalKey("mymodule", c); err != nil {
        fmt.Printf("failed to reload mymodule config: %v\n", err)
        return err
    }
    fmt.Printf("mymodule config reload success, config: %+v\n", c)
    // 重新初始化以应用新配置
    Reinit()
    return nil
}
```

### 热更新机制

- 配置文件变更时自动触发热更新
- 只更新实现了 `HotReloadable` 接口的模块
- 基础配置会自动更新
- 热更新是安全的，不会影响正在运行的请求

**注意事项：**

- 只有标记为"安全"的配置项才应该支持热更新
- 涉及连接、认证等关键配置不建议热更新
- 热更新回调函数应该是幂等的

## 完整示例

### 1. 定义模块配置

```go
package mymodule

import (
    "fmt"
    "github.com/jessewkun/gocommon/config"
    "github.com/spf13/viper"
)

type Config struct {
    Host    string `mapstructure:"host" json:"host"`
    Port    int    `mapstructure:"port" json:"port"`
    Timeout int    `mapstructure:"timeout" json:"timeout"`
}

var Cfg = &Config{
    Host:    "localhost",
    Port:    8080,
    Timeout: 5,
}

var client *Client

func init() {
    // 注册配置
    config.Register("mymodule", Cfg)
    // 注册初始化回调，依赖 config 和 log
    config.RegisterCallback("mymodule", Init, "config", "log")
}

// Init 初始化函数
func Init() error {
    client = NewClient(Cfg.Host, Cfg.Port)
    fmt.Printf("mymodule initialized: %+v\n", Cfg)
    return nil
}

// Reload 热更新支持
func (c *Config) Reload(v *viper.Viper) {
    if err := v.UnmarshalKey("mymodule", c); err != nil {
        fmt.Printf("failed to reload mymodule config: %v\n", err)
        return
    }
    fmt.Printf("mymodule config reload success, config: %+v\n", c)
    // 重新初始化以应用新配置
    Init()
}
```

### 2. 配置文件

**config.toml：**

```toml
mode = "debug"
port = ":8000"
app_name = "MyService"
domain = "http://localhost:8000"

[mymodule]
host = "localhost"
port = 8080
timeout = 5
```

### 3. 应用启动

```go
package main

import (
    "log"
    "github.com/jessewkun/gocommon/config"
    _ "github.com/yourproject/mymodule" // 导入模块以触发 init()
)

func main() {
    // 加载配置（会自动加载所有注册的模块配置）
    cfg, err := config.Init("./config.toml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 使用配置
    log.Printf("App started: %s on %s", cfg.AppName, cfg.Port)

    // 模块配置已自动加载和初始化
    // mymodule.Cfg 和 mymodule.client 已就绪
}
```

## 配置文件格式

### TOML 格式（推荐）

TOML 格式可读性好，适合复杂配置：

```toml
mode = "debug"
port = ":8000"

[log]
path = "./logs/app.log"
max_size = 100

[mysql.default]
dsn = ["user:password@tcp(localhost:3306)/dbname"]
max_conn = 100
```

### JSON 格式

JSON 格式适合程序生成和 API 交互：

```json
{
  "mode": "debug",
  "port": ":8000",
  "log": {
    "path": "./logs/app.log",
    "max_size": 100
  },
  "mysql": {
    "default": {
      "dsn": ["user:password@tcp(localhost:3306)/dbname"],
      "max_conn": 100
    }
  }
}
```

### YAML 格式

YAML 格式简洁，适合人工编辑：

```yaml
mode: debug
port: ":8000"

log:
  path: "./logs/app.log"
  max_size: 100

mysql:
  default:
    dsn:
      - "user:password@tcp(localhost:3306)/dbname"
    max_conn: 100
```

## 高级功能

### 配置验证

在初始化回调函数中进行配置验证：

```go
func Init() error {
    if Cfg.Host == "" {
        return fmt.Errorf("host is required")
    }
    if Cfg.Port <= 0 {
        return fmt.Errorf("port must be positive")
    }
    // 初始化逻辑...
    return nil
}
```

### 默认值处理

配置结构体字段的零值可以作为默认值：

```go
var Cfg = &Config{
    Host:    "localhost", // 默认值
    Port:    8080,        // 默认值
    Timeout: 5,           // 默认值
}
```

如果配置文件中没有对应字段，将使用默认值。

### 配置缺失处理

如果配置文件中没有某个模块的配置，会输出警告但不会报错：

```
config: key 'mymodule' not found in config file, module will use default values
```

### 错误处理

**初始化错误：**

- 如果 `log` 模块初始化失败，会中断整个初始化流程
- 其他模块初始化失败只会打印错误，继续执行

**热更新错误：**

- 热更新失败不会影响当前运行的服务
- 错误信息会打印到控制台

## 注意事项

1. **回调函数幂等性**：确保初始化回调函数是幂等的，因为热更新可能会多次调用
2. **依赖顺序**：正确声明模块依赖关系，避免循环依赖
3. **配置安全**：只有安全的配置项才支持热更新，关键配置（如数据库连接）不建议热更新
4. **线程安全**：配置读取是线程安全的，但模块需要自行保证配置使用的线程安全
5. **配置文件路径**：使用绝对路径或相对于工作目录的路径
6. **配置格式**：确保配置文件格式正确，否则会导致加载失败

## 测试

运行测试：

```sh
go test ./config -v
```

测试覆盖：

- TOML 格式配置加载
- JSON 格式配置加载
- YAML 格式配置加载
- 基础配置验证

## 示例配置文件

完整示例配置文件请参考：

- `config.toml.example`：TOML 格式示例
- `config.json.example`：JSON 格式示例

## 与其他模块集成

配置模块是其他所有模块的基础，各模块通过以下方式集成：

1. **注册配置**：在 `init()` 中调用 `config.Register()`
2. **注册回调**：在 `init()` 中调用 `config.RegisterCallback()`
3. **实现热更新**：实现 `HotReloadable` 接口（可选）

## 最佳实践

1. **配置结构设计**：
   - 使用有意义的字段名
   - 提供合理的默认值
   - 使用 `mapstructure` 标签支持配置映射

2. **初始化回调**：
   - 在回调中进行配置验证
   - 执行必要的初始化逻辑
   - 确保回调函数是幂等的

3. **热更新实现**：
   - 只对安全的配置项支持热更新
   - 在 `Reload` 中重新初始化相关组件
   - 添加适当的日志输出

4. **依赖管理**：
   - 明确声明模块依赖
   - 避免循环依赖
   - 合理使用注入器函数处理复杂依赖
