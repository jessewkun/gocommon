// Package gemini an provider for google gemini
package gemini

import (
	"time"
)

// Config for the Gemini provider
type Config struct {
	APIKey  string        `mapstructure:"api_key"`
	APIURL  string        `mapstructure:"api_url"` // e.g., https://generativelanguage.googleapis.com
	Timeout time.Duration `mapstructure:"timeout"`
}

// GeminiPart is a part of a GeminiContent
// Can contain either text or inline binary data (e.g., images)
type GeminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// InlineData represents inline binary data (e.g., images) for multimodal input
type InlineData struct {
	MimeType string `json:"mimeType"` // e.g., "image/png", "image/jpeg"
	Data     string `json:"data"`     // Base64-encoded binary data
}

// GeminiContent is a single message in the Gemini chat history
type GeminiContent struct {
	Role  string       `json:"role"` // "user" or "model"
	Parts []GeminiPart `json:"parts"`
}

// GeminiCandidate is a single response candidate from the Gemini API
type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

// PromptFeedback contains feedback about the prompt
type PromptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// SafetyRating represents the safety rating of a response
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiChatResponse is the response from the Gemini Chat API
type GeminiChatResponse struct {
	Candidates     []GeminiCandidate `json:"candidates"`
	PromptFeedback *PromptFeedback   `json:"promptFeedback,omitempty"`
}

// GeminiEmbeddingRequest is the request for the Gemini Embedding API
type GeminiEmbeddingRequest struct {
	Model   string        `json:"model"`
	Content GeminiContent `json:"content"`
}

// EmbeddingValue holds the embedding vector
type EmbeddingValue struct {
	Value []float32 `json:"value"`
}

// GeminiEmbeddingResponse is the response from the Gemini Embedding API
type GeminiEmbeddingResponse struct {
	Embedding EmbeddingValue `json:"embedding"`
}

// GenerationConfig controls the generation of the response
type GenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GeminiChatRequest is the request to the Gemini Chat API
type GeminiChatRequest struct {
	Contents          []GeminiContent   `json:"contents"`
	SystemInstruction *GeminiContent    `json:"system_instruction,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
}

// --- Stream-specific types ---

// GeminiStreamCandidate is a candidate in a streaming response
type GeminiStreamCandidate struct {
	Content       GeminiContent  `json:"content"`
	FinishReason  string         `json:"finishReason,omitempty"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiStreamUsageMetadata contains token count for streaming
type GeminiStreamUsageMetadata struct {
	PromptTokenCount int `json:"promptTokenCount"`
	TotalTokenCount  int `json:"totalTokenCount"`
}

// GeminiStreamResponse is a single chunk in a streaming response
type GeminiStreamResponse struct {
	Candidates    []GeminiStreamCandidate    `json:"candidates"`
	UsageMetadata *GeminiStreamUsageMetadata `json:"usageMetadata,omitempty"`
}
