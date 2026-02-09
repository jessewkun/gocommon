// Package llm 提供大模型调用的 Provider 抽象，业务只依赖本包，通过名称+配置获取 Provider，使用统一 ChatRequest 调用。
package llm

import "context"

// Provider 大模型提供方的基础接口，只包含名称
type Provider interface {
	Name() string
}

// Chatter 接口定义了聊天功能
type Chatter interface {
	Provider
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest, callback func(chunk string) error) (*ChatResponse, error)
}

// Embedder 接口定义了向量化功能
type Embedder interface {
	Provider
	CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
}

// 未来可以继续扩展
// type ImageGenerator interface { ... }
