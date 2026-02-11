package llm

import "fmt"

// ContentString 从 Message.Content 提取字符串；仅当 Content 为 string 时返回有效文本，否则返回空串
// 用于仅支持纯文本的 Provider（如 gemini）
func ContentString(c interface{}) string {
	if c == nil {
		return ""
	}
	if s, ok := c.(string); ok {
		return s
	}
	return fmt.Sprint(c)
}

// Message 单条对话消息，各 Provider 通用
// Content 为 string 时表示纯文本；为 []interface{} 时表示多模态（如 [{"type":"text","text":"..."},{"type":"image_url","image_url":{"url":"..."}}]）
type Message struct {
	Role    string      // system / user / assistant
	Content interface{} // 消息内容：string 或 []interface{}（多模态）
}

// Usage 统计 API 调用过程中的 token 使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatRequest 统一对话请求
type ChatRequest struct {
	Model            string                   // 模型名
	Messages         []Message                // 对话历史
	Temperature      float64                  // 温度
	TopP             *float64                 // top_p
	TopK             *int                     // top_k（部分 Provider 支持）
	MaxTokens        *int                     // max_tokens / max_output_tokens
	Stop             []string                 // stop
	PresencePenalty  *float64                 // presence_penalty
	FrequencyPenalty *float64                 // frequency_penalty
	N                *int                     // n / candidate count
	User             string                   // user
	Seed             *int                     // seed
	LogitBias        map[string]int           // logit_bias
	Tools            []map[string]interface{} // tools（OpenAI 兼容）
	ToolChoice       interface{}              // tool_choice（OpenAI 兼容）
	ResponseFormat   string                   // 如 "json_object" 或 "text"
	Reasoning        *ReasoningConfig         // OpenRouter reasoning 参数
}

// ReasoningConfig 对应 OpenRouter reasoning 参数
// 参考：https://openrouter.ai/docs/guides/best-practices/reasoning-tokens
// effort 与 max_tokens 二选一；exclude / enabled 可选
type ReasoningConfig struct {
	Effort    string `json:"effort,omitempty"`     // "xhigh" | "high" | "medium" | "low" | "minimal" | "none"
	MaxTokens *int   `json:"max_tokens,omitempty"` // reasoning token 上限
	Exclude   *bool  `json:"exclude,omitempty"`    // 是否从响应中排除 reasoning
	Enabled   *bool  `json:"enabled,omitempty"`    // 是否启用 reasoning（默认根据 effort/max_tokens 推断）
}

// ChatResponse 统一对话响应
type ChatResponse struct {
	Content      string // 模型返回的主要内容
	FinishReason string // 结束原因，如 "stop", "length"
	Usage        Usage  // Token 使用情况
	RawResponse  []byte `json:"-"` // 原始响应体，用于调试或特殊用途
}

// EmbeddingRequest 统一 Embedding 请求
type EmbeddingRequest struct {
	Model string   // 模型名
	Input []string // 需要被向量化的文本列表
}

// Embedding 单个文本的向量结果
type Embedding struct {
	Index  int       `json:"index"`
	Vector []float32 `json:"embedding"`
	Object string    `json:"object"`
}

// EmbeddingResponse 统一 Embedding 响应
type EmbeddingResponse struct {
	Data        []Embedding `json:"data"`
	Model       string      `json:"model"`
	Usage       Usage       `json:"usage"`
	RawResponse []byte      `json:"-"` // 原始响应体
}
