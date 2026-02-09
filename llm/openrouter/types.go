package openrouter

import (
	"time"

	"github.com/jessewkun/gocommon/llm"
)

// Config OpenRouter/OpenAI 兼容 API 的配置
type Config struct {
	APIURL  string        `mapstructure:"api_url"` // API 地址, e.g., "https://openrouter.ai/api/v1"
	APIKey  string        `mapstructure:"api_key"` // API Key
	Timeout time.Duration `mapstructure:"timeout"` // 请求超时
}

// ---- 下面是 OpenAI 兼容的 API 响应结构 ----

// OpenAIChatChoice is a single choice in a non-streaming chat response
type OpenAIChatChoice struct {
	Index        int         `json:"index"`
	Message      llm.Message `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// OpenAIResponse 是非流式响应的结构
type OpenAIResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenAIChatChoice `json:"choices"`
	Usage   llm.Usage          `json:"usage"`
	Error   *APIError          `json:"error,omitempty"`
}

// OpenAIStreamChoice is a single choice in a streaming response
type OpenAIStreamChoice struct {
	Index int `json:"index"`
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

// OpenAIStreamResponse 是流式响应中单个 data 块的结构
type OpenAIStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
	Error   *APIError            `json:"error,omitempty"`
}

// APIError 封装了 API 可能返回的错误信息
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

// OpenAIEmbeddingData holds the data for a single embedding
type OpenAIEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// OpenAIEmbeddingResponse 是 Embedding API 的响应结构
type OpenAIEmbeddingResponse struct {
	Object string                `json:"object"`
	Data   []OpenAIEmbeddingData `json:"data"`
	Model  string                `json:"model"`
	Usage  llm.Usage             `json:"usage"`
	Error  *APIError             `json:"error,omitempty"`
}
