// Package openai an provider for openai
package openai

import (
	"time"

	"github.com/jessewkun/gocommon/llm"
)

// Config for the OpenAI provider
type Config struct {
	APIURL  string        `mapstructure:"api_url"` // API Base URL, defaults to "https://api.openai.com/v1"
	APIKey  string        `mapstructure:"api_key"` // API Key
	Timeout time.Duration `mapstructure:"timeout"` // Request timeout
}

// ---- OpenAI API compatible response structures ----

// OpenAIChatChoice is a single choice in a non-streaming chat response
type OpenAIChatChoice struct {
	Index        int         `json:"index"`
	Message      llm.Message `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// OpenAIResponse is the response structure for non-streaming chat completions
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

// OpenAIStreamResponse is the structure of a single data chunk in a streaming response
type OpenAIStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
	Error   *APIError            `json:"error,omitempty"`
}

// APIError encapsulates error information from the API
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

// OpenAIEmbeddingResponse is the response from the OpenAI Embedding API
type OpenAIEmbeddingResponse struct {
	Object string                `json:"object"`
	Data   []OpenAIEmbeddingData `json:"data"`
	Model  string                `json:"model"`
	Usage  llm.Usage             `json:"usage"`
	Error  *APIError             `json:"error,omitempty"`
}
