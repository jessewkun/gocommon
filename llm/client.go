package llm

import (
	"context"
	"fmt"
)

// Client 是与 LLM 服务交互的统一客户端
type Client struct {
	provider Provider
}

// NewClient 使用指定的 Provider 创建一个新的客户端
func NewClient(name string, config interface{}) (*Client, error) {
	provider, err := NewProvider(name, config)
	if err != nil {
		return nil, err
	}
	return &Client{provider: provider}, nil
}

// Chat 执行一次聊天请求 (非流式)
// 如果 Provider 不支持聊天，将返回错误
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	chatter, ok := c.provider.(Chatter)
	if !ok {
		return nil, fmt.Errorf("llm: provider %q does not support chat", c.provider.Name())
	}
	return chatter.Chat(ctx, req)
}

// ChatStream 执行一次聊天请求 (流式)
// 如果 Provider 不支持聊天，将返回错误
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest, callback func(chunk string) error) (*ChatResponse, error) {
	chatter, ok := c.provider.(Chatter)
	if !ok {
		return nil, fmt.Errorf("llm: provider %q does not support chat", c.provider.Name())
	}
	return chatter.ChatStream(ctx, req, callback)
}

// CreateEmbeddings 执行一次向量化请求
// 如果 Provider 不支持向量化，将返回错误
func (c *Client) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	embedder, ok := c.provider.(Embedder)
	if !ok {
		return nil, fmt.Errorf("llm: provider %q does not support embeddings", c.provider.Name())
	}
	return embedder.CreateEmbeddings(ctx, req)
}

// ProviderName 返回当前客户端使用的 Provider 名称
func (c *Client) ProviderName() string {
	return c.provider.Name()
}
