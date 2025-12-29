# 告警模块

## 功能简介

告警模块提供统一的告警接口，支持同时发送到多个配置的渠道。目前支持 Bark（iOS 推送）和飞书机器人两种告警渠道。

- ✅ 统一告警接口，支持同时发送到多个渠道
- ✅ 支持 Bark iOS 推送
- ✅ 支持飞书群机器人告警
- ✅ 并发发送，提高告警可靠性
- ✅ 完善的错误处理机制，单个渠道失败不影响其他渠道
- ✅ 支持配置热更新
- ✅ 自动重试机制（最多重试 2 次）
- ✅ 支持 logger.Alerter 接口适配

## 依赖

- `github.com/jessewkun/gocommon/http`：HTTP 客户端模块
- `github.com/jessewkun/gocommon/config`：配置管理模块

## 配置说明

### Config 配置结构

```go
type Config struct {
    Bark    *Bark   `mapstructure:"bark" json:"bark"`       // Bark 配置
    Feishu  *Feishu `mapstructure:"feishu" json:"feishu"`    // 飞书配置
    Timeout int     `mapstructure:"timeout" json:"timeout"` // 请求超时时间（秒），默认 5 秒
}
```

### Bark 配置

```go
type Bark struct {
    BarkIds []string `mapstructure:"bark_ids" json:"bark_ids"` // Bark 设备 ID 列表
}
```

### Feishu 配置

```go
type Feishu struct {
    WebhookURL string `mapstructure:"webhook_url" json:"webhook_url"` // 飞书机器人 Webhook URL
    Secret     string `mapstructure:"secret" json:"secret"`          // 飞书机器人 Secret（可选）
}
```

### 配置示例

**`config.toml` 中的配置：**

```toml
[alarm]
timeout = 5

[alarm.bark]
bark_ids = ["jT64URJj8b6Fp9Y3nVKJiP", "another_device_id"]

[alarm.feishu]
webhook_url = "https://open.feishu.cn/open-apis/bot/v2/hook/your_token"
secret = "your_secret_key"
```

**`config.json` 中的配置：**

```json
{
  "alarm": {
    "timeout": 5,
    "bark": {
      "bark_ids": ["jT64URJj8b6Fp9Y3nVKJiP", "another_device_id"]
    },
    "feishu": {
      "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/your_token",
      "secret": "your_secret_key"
    }
  }
}
```

## 基本使用

### 1. 初始化

应用启动时，`config` 模块会自动调用 `alarm.Init()` 进行初始化。你也可以手动调用：

```go
import "github.com/jessewkun/gocommon/alarm"

if err := alarm.Init(); err != nil {
    log.Fatalf("Failed to initialize alarm: %v", err)
}
```

### 2. 发送告警

使用统一的 `SendAlarm` 接口发送告警：

```go
import (
    "context"
    "github.com/jessewkun/gocommon/alarm"
)

ctx := context.Background()
title := "系统告警"
content := []string{
    "- 告警内容第一行",
    "- 告警内容第二行",
    "- 告警内容第三行",
}

if err := alarm.SendAlarm(ctx, title, content); err != nil {
    log.Printf("Failed to send alarm: %v", err)
}
```

### 3. 使用 Sender 适配器

`Sender` 实现了 `logger.Alerter` 接口，可以直接用于日志模块的告警：

```go
import (
    "github.com/jessewkun/gocommon/alarm"
    "github.com/jessewkun/gocommon/logger"
)

// 创建 Sender 实例
sender := &alarm.Sender{}

// 发送告警（实现 logger.Alerter 接口）
err := sender.Send(ctx, "日志告警", []string{"错误信息", "堆栈信息"})
if err != nil {
    log.Printf("Failed to send alarm: %v", err)
}
```

## 支持的告警渠道

### Bark（iOS 推送）

Bark 是一个 iOS 推送服务，支持向多个设备发送推送通知。

**配置说明：**

1. 在 iOS 设备上安装 Bark 应用
2. 获取设备 ID（Bark ID）
3. 在配置文件中添加 `bark_ids` 列表

**特性：**

- 支持向多个设备并发发送
- 自动 URL 编码标题和内容
- 单个设备失败不影响其他设备

**使用示例：**

```go
// 直接使用 Bark 发送
bark := &alarm.Bark{
    BarkIds: []string{"jT64URJj8b6Fp9Y3nVKJiP"},
}

err := bark.Send(ctx, "Bark 测试", []string{"这是一条测试消息"})
if err != nil {
    log.Printf("Failed to send Bark: %v", err)
}
```

### 飞书机器人

飞书机器人支持向飞书群发送富文本消息，适合团队协作场景。

**配置说明：**

1. 在飞书群中添加自定义机器人
2. 获取 Webhook URL
3. 可选：配置 Secret 用于签名验证
4. 在配置文件中添加 `webhook_url` 和 `secret`

**特性：**

- 支持富文本消息格式
- 支持签名验证（可选）
- 自动生成时间戳和签名

**使用示例：**

```go
// 直接使用飞书发送
feishu := &alarm.Feishu{
    WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/your_token",
    Secret:     "your_secret_key",
}

err := feishu.Send(ctx, "飞书告警", []string{
    "- 告警类型：系统异常",
    "- 告警时间：2025-01-01 12:00:00",
    "- 告警详情：数据库连接失败",
})
if err != nil {
    log.Printf("Failed to send Feishu: %v", err)
}
```

## 高级功能

### 并发发送

`SendAlarm` 接口会自动并发发送到所有配置的渠道，提高告警的可靠性：

```go
// 同时发送到 Bark 和飞书
err := alarm.SendAlarm(ctx, "系统告警", []string{"告警内容"})
// 如果 Bark 失败但飞书成功，会返回部分失败的错误信息
// 如果所有渠道都成功，返回 nil
```

### 错误处理

告警模块具有完善的错误处理机制：

- 单个渠道失败不影响其他渠道
- 返回所有失败渠道的错误信息
- 自动记录错误日志

```go
err := alarm.SendAlarm(ctx, "告警", []string{"内容"})
if err != nil {
    // 错误信息格式：failed to send alarm to some channels: [bark failed: ...; feishu failed: ...]
    log.Printf("部分渠道发送失败: %v", err)
}
```

### 配置热更新

告警模块支持配置热更新，无需重启服务：

```go
// 配置变更后，会自动调用 Reload 方法
// 重新初始化 HTTP 客户端以应用新的超时设置
```

**注意事项：**

- 所有配置项都被认为是安全的，可以进行热更新
- 热更新后会自动重新初始化 HTTP 客户端

### 重试机制

告警模块内置重试机制：

- 默认最多重试 2 次（`MaxRetry = 2`）
- 使用 `@http` 模块的重试功能
- 自动处理网络错误和超时

## 注意事项

1. **配置检查**：如果没有配置任何告警渠道，`SendAlarm` 会返回错误
2. **超时设置**：建议根据网络环境调整 `timeout` 配置，默认 5 秒
3. **日志循环**：告警模块不记录自己的请求日志，避免循环告警
4. **Bark ID 获取**：需要在 iOS 设备上安装 Bark 应用并获取设备 ID
5. **飞书 Webhook**：确保 Webhook URL 格式正确，包含完整的 token
6. **飞书签名**：如果配置了 Secret，会自动生成签名进行验证
7. **并发安全**：所有发送操作都是并发安全的

## 测试

运行测试：

```sh
go test ./alarm -v
```

测试说明：

- 测试代码中使用了示例配置，实际运行时需要替换为真实的配置
- 网络错误（如 404）是可以接受的，只要不是配置错误即可
- 测试会验证配置验证逻辑和错误处理机制

## 示例代码

完整的使用示例请参考本目录下的测试文件：

- `alarm_test.go`：统一接口测试
- `bark_test.go`：Bark 渠道测试
- `feishu_test.go`：飞书渠道测试

## 与日志模块集成

告警模块可以与日志模块无缝集成：

```go
import (
    "github.com/jessewkun/gocommon/alarm"
    "github.com/jessewkun/gocommon/logger"
)

// 创建 Sender 实例
sender := &alarm.Sender{}

// 配置日志模块使用告警
logger.SetAlerter(sender)

// 当日志级别达到告警阈值时，会自动发送告警
logger.Error("系统异常", logger.Field("error", err))
```

## 依赖模块

- **config**：配置管理和热更新
- **http**：HTTP 客户端，支持重试和超时
- **logger**：日志模块（可选，用于集成告警）
