# LLM Provider 模块

## 简介

本项目提供了一套**统一的、可扩展的大模型（LLM）调用客户端**。
业务层只需通过名称和配置即可获取 `llm.Client` 实例，并使用统一的请求和响应结构进行 Chat (聊天) 和 Embedding (向量化) 调用，无需关心底层具体的模型服务实现（如 OpenAI, OpenRouter, Google Gemini 等）。

## 核心设计

1.  **接口分离**：定义了 `llm.Provider` (基础接口)、`llm.Chatter` (聊天能力接口) 和 `llm.Embedder` (向量化能力接口)。具体模型实现可以按需实现一个或多个接口。
2.  **统一客户端**：`llm.Client` 作为业务层的唯一入口，它内部持有一个 `llm.Provider` 实例，并根据 Provider 是否实现 `llm.Chatter` 或 `llm.Embedder` 接口，提供类型安全的 `Chat()`、`ChatStream()` 和 `CreateEmbeddings()` 方法。
3.  **丰富数据模型**：请求（`llm.ChatRequest`, `llm.EmbeddingRequest`）和响应（`llm.ChatResponse`, `llm.EmbeddingResponse`）都使用结构体，包含 Token 用量、结束原因等更多元数据。
4.  **多模态支持**：`llm.Message.Content` 支持 `string`（纯文本）或 `[]interface{}`（多模态数组），所有 Provider 均支持图片输入。
5.  **注册机制**：采用 Go 的 `init()` 和空白导入 (`import _ "..."`) 模式，实现 Provider 的动态注册和加载，保证核心 `llm` 包与具体实现彻底解耦。

## 使用方式

### 1. 导入并注册 Provider

需要使用哪个 Provider，在 `main` 函数或初始化处做一次**空白导入**以触发其 `init()` 函数进行注册：

```go
import (
    "github.com/jessewkun/gocommon/llm"
    _ "github.com/jessewkun/gocommon/llm/openai"     // 注册 OpenAI Provider
    _ "github.com/jessewkun/gocommon/llm/openrouter" // 注册 OpenRouter Provider
    _ "github.com/jessewkun/gocommon/llm/gemini"     // 注册 Gemini Provider
)
```

### 2. 创建 LLM 客户端

使用 Provider 的名称 + 配置来创建 `llm.Client` 实例。配置通常是一个 `map[string]interface{}`，也可以是具体 Provider 定义的配置结构体。

```go
// 示例：创建 openrouter 客户端
client, err := llm.NewClient("openrouter", map[string]interface{}{
    "api_url":     "https://openrouter.ai/api/v1",
    "api_key":     "your-openrouter-api-key",
    "timeout": 60 * time.Second, // time.Duration 类型
})
if err != nil {
    log.Fatalf("创建 LLM 客户端失败: %v", err)
}
fmt.Printf("LLM 客户端使用 Provider: %s\n", client.ProviderName())
```

### 3. 调用 Chat (聊天) 功能

#### 特殊说明：System Prompt (系统提示)

对于需要设置系统角色的场景 (例如 `gemini` 或 `openai`)，请将 `system` 角色的消息放在 `Messages` 列表的**第一位**。底层的 Provider 会自动将其转换为对应 API 的格式（例如 Gemini 的 `system_instruction`）。

```go
// 带 system prompt 的请求
chatReq := &llm.ChatRequest{
    Model: "gemini-pro",
    Messages: []llm.Message{
        {Role: "system", Content: "你是一个专业的翻译官，将所有内容翻译成英文。"},
        {Role: "user", Content: "你好，请介绍一下你自己。"},
    },
    Temperature: 0.7,
}
```

#### 非流式调用

```go
resp, err := client.Chat(ctx, chatReq)
if err != nil {
    log.Fatalf("Chat 调用失败: %v", err)
}
fmt.Println("===== Chat Response =====")
fmt.Printf("内容: %s\n", resp.Content)
fmt.Printf("结束原因: %s\n", resp.FinishReason)
fmt.Printf("Token 用量: %+v\n", resp.Usage)
// fmt.Printf("原始响应: %s\n", string(resp.RawResponse)) // 调试用
```

#### 流式调用

```go
fmt.Println("\n===== ChatStream Response =====")
fullResp, err := client.ChatStream(ctx, chatReq, func(chunk string) error {
    fmt.Print(chunk) // 实时打印流式输出
    return nil
})
if err != nil {
    log.Fatalf("ChatStream 调用失败: %v", err)
}
fmt.Println("\n--- 流式结束 ---")
fmt.Printf("完整内容: %s\n", fullResp.Content)
fmt.Printf("结束原因: %s\n", fullResp.FinishReason)
// 流式调用通常不直接返回 Usage，需要 Provider 特殊处理或通过 API 头部获取
```

#### 多模态调用 (图片输入)

所有 Provider (openai, openrouter, gemini) 均支持多模态输入。使用 `[]interface{}` 类型的 `Content` 字段传递文本和图片：

```go
// 多模态请求：文本 + 图片
chatReq := &llm.ChatRequest{
    Model: "gemini-pro-vision", // 或 gpt-4o, claude-3-opus 等支持多模态的模型
    Messages: []llm.Message{
        {
            Role: "user",
            Content: []interface{}{
                map[string]interface{}{"type": "text", "text": "这张图片里有什么？"},
                map[string]interface{}{
                    "type": "image_url",
                    "image_url": map[string]interface{}{
                        "url": "data:image/png;base64,iVBORw0KGgo...", // Base64 编码的图片
                    },
                },
            },
        },
    },
}

resp, err := client.Chat(ctx, chatReq)
```

**Content 格式说明：**
-   纯文本：`Content: "你好"` (string 类型)
-   多模态：`Content: []interface{}{...}` (数组类型)，支持以下元素：
    -   文本：`{"type": "text", "text": "..."}`
    -   图片：`{"type": "image_url", "image_url": {"url": "data:<mime>;base64,<data>"}}`

### 4. 调用 Embedding (向量化) 功能

如果所选 Provider 实现了 `llm.Embedder` 接口，则可以进行向量化调用：

```go
embeddingReq := &llm.EmbeddingRequest{
    Model: "text-embedding-ada-002", // 使用所选 Provider 支持的 embedding 模型
    Input: []string{"Go语言编程", "分布式系统设计"},
}

embResp, err := client.CreateEmbeddings(ctx, embeddingReq)
if err != nil {
    log.Fatalf("Embedding 调用失败: %v", err)
}
fmt.Println("\n===== Embedding Response =====")
fmt.Printf("向量数量: %d\n", len(embResp.Data))
if len(embResp.Data) > 0 {
    fmt.Printf("第一个向量维度: %d\n", len(embResp.Data[0].Vector))
}
fmt.Printf("Token 用量: %+v\n", embResp.Usage)
```

## 支持的提供商 (Supported Providers)

-   每个具体 Provider (如 `openrouter`) 都有其特定的配置项。
-   通常，`map[string]interface{}` 方式可用于所有 Provider。
-   若业务已引用具体 Provider 包（如 `_ "github.com/.../llm/openrouter"`），也可以直接传递其定义的 `Config` 结构体。

### OpenAI

-   **名称**: `openai`
-   **能力**: Chat, ChatStream, Embedding, **多模态**
-   **配置项**:
    -   `api_key`: (必须) OpenAI API Key。
    -   `api_url`: (可选) API 地址，默认为 `https://api.openai.com/v1`。
    -   `timeout`: (可选) 请求超时，`time.Duration` 类型。

### OpenRouter

-   **名称**: `openrouter`
-   **说明**: OpenRouter 是一个兼容 OpenAI API 的路由服务。
-   **能力**: Chat, ChatStream, Embedding, **多模态**
-   **配置项**:
    -   `api_key`: (必须) OpenRouter API Key。
    -   `api_url`: (可选) API 地址，默认为 `https://openrouter.ai/api/v1`。
    -   `timeout`: (可选) 请求超时，`time.Duration` 类型。

### Google Gemini

-   **名称**: `gemini`
-   **能力**: Chat, ChatStream, Embedding, **多模态**
-   **配置项**:
    -   `api_key`: (必须) Gemini API Key。
    -   `api_url`: (可选) API 地址，默认为 `https://generativelanguage.googleapis.com`。
    -   `timeout`: (可选) 请求超时，`time.Duration` 类型。

## 扩展其他模型

若要集成一个新的大模型（例如 `anthropic`）：

1.  在 `llm/anthropic` 目录下，实现 `llm.Chatter` 和/或 `llm.Embedder` 接口。
2.  在 `llm/anthropic/register.go` 的 `init()` 函数中，调用 `llm.Register("anthropic", factoryFunction)` 来注册你的 Provider 工厂。
3.  业务侧通过空白导入 `_ "github.com/.../llm/anthropic"` 来加载，并使用 `llm.NewClient("anthropic", config)` 创建客户端。
