// Package gemini an provider for google gemini
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	xhttp "github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/llm"
	"github.com/jessewkun/gocommon/logger"
)

const logTag = "LLM_GEMINI"

// Provider implements llm.Chatter and llm.Embedder for Google Gemini
type Provider struct {
	client *xhttp.Client
	cfg    Config
}

// NewProvider creates a new Gemini Provider
func NewProvider(client *xhttp.Client, cfg Config) *Provider {
	return &Provider{client: client, cfg: cfg}
}

// Name implements llm.Provider
func (p *Provider) Name() string {
	return providerName
}

// --- Chatter Implementation ---

// Chat implements llm.Chatter for non-streaming chat
func (p *Provider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	geminiReq, err := p.toGeminiChatRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshalling chat request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.cfg.APIURL, req.Model, p.cfg.APIKey)

	resp, err := p.client.Post(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: map[string]string{"Content-Type": "application/json"},
		Timeout: p.cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: chat api call failed: %w", err)
	}

	var geminiResp GeminiChatResponse
	if err := json.Unmarshal(resp.Body, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: unmarshalling chat response: %w (body: %s)", err, string(resp.Body))
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: no candidates returned in response")
	}

	return p.toLLMChatResponse(&geminiResp), nil
}

// ChatStream implements llm.Chatter for streaming chat
func (p *Provider) ChatStream(ctx context.Context, req *llm.ChatRequest, callback func(chunk string) error) (*llm.ChatResponse, error) {
	geminiReq, err := p.toGeminiChatRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshalling stream request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.cfg.APIURL, req.Model, p.cfg.APIKey)

	var fullContent strings.Builder
	var finalResponse llm.ChatResponse

	timeout := p.cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	err = p.client.PostStream(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: map[string]string{"Content-Type": "application/json"},
		Timeout: timeout,
	}, func(line []byte) error {
		if !bytes.HasPrefix(line, []byte("data: ")) {
			return nil
		}

		dataStr := bytes.TrimPrefix(line, []byte("data: "))
		if len(dataStr) == 0 {
			return nil
		}

		var streamResp GeminiStreamResponse
		if err := json.Unmarshal(dataStr, &streamResp); err != nil {
			logger.WarnWithField(ctx, logTag, "failed to parse gemini stream line", map[string]interface{}{
				"line":  string(line),
				"error": err.Error(),
			})
			return nil // Continue to next line
		}

		if len(streamResp.Candidates) > 0 {
			candidate := streamResp.Candidates[0]
			// Check for content
			if len(candidate.Content.Parts) > 0 && candidate.Content.Parts[0].Text != "" {
				chunk := candidate.Content.Parts[0].Text
				fullContent.WriteString(chunk)
				if err := callback(chunk); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
			// Check for finish reason
			if candidate.FinishReason != "" {
				finalResponse.FinishReason = candidate.FinishReason
			}
		}
		if streamResp.UsageMetadata != nil {
			finalResponse.Usage.PromptTokens = streamResp.UsageMetadata.PromptTokenCount
			finalResponse.Usage.TotalTokens = streamResp.UsageMetadata.TotalTokenCount
			// CompletionTokens can be derived if needed
			finalResponse.Usage.CompletionTokens = streamResp.UsageMetadata.TotalTokenCount - streamResp.UsageMetadata.PromptTokenCount
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("gemini: stream failed: %w", err)
	}

	finalResponse.Content = fullContent.String()
	return &finalResponse, nil
}

// --- Embedder Implementation ---

// CreateEmbeddings implements llm.Embedder
func (p *Provider) CreateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	if len(req.Input) != 1 {
		// Gemini API v1beta currently supports one input string at a time for embedding.
		// For multiple inputs, batching needs to be implemented here.
		// This example will handle only the first input for simplicity.
		logger.Warn(ctx, logTag, "Gemini provider currently handles only the first input for embedding.")
		if len(req.Input) == 0 {
			return nil, fmt.Errorf("gemini: embedding input cannot be empty")
		}
	}

	geminiReq := GeminiEmbeddingRequest{
		Model: "models/" + req.Model,
		Content: GeminiContent{
			Role:  "user",
			Parts: []GeminiPart{{Text: req.Input[0]}},
		},
	}
	bodyBytes, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshalling embedding request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s", p.cfg.APIURL, req.Model, p.cfg.APIKey)

	resp, err := p.client.Post(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: map[string]string{"Content-Type": "application/json"},
		Timeout: p.cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: embedding api call failed: %w", err)
	}

	var geminiResp GeminiEmbeddingResponse
	if err := json.Unmarshal(resp.Body, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: unmarshalling embedding response: %w (body: %s)", err, string(resp.Body))
	}

	return &llm.EmbeddingResponse{
		Data: []llm.Embedding{
			{
				Index:  0,
				Vector: geminiResp.Embedding.Value,
				Object: "embedding",
			},
		},
		Model: req.Model,
		// Gemini embedding API doesn't provide token usage details in the response
	}, nil
}

// --- Helper Functions ---

func (p *Provider) toGeminiChatRequest(ctx context.Context, req *llm.ChatRequest) (*GeminiChatRequest, error) {
	geminiReq := &GeminiChatRequest{
		GenerationConfig: &GenerationConfig{
			Temperature: req.Temperature,
		},
	}
	if req.TopP != nil {
		geminiReq.GenerationConfig.TopP = req.TopP
	}
	if req.TopK != nil {
		geminiReq.GenerationConfig.TopK = req.TopK
	}
	if req.MaxTokens != nil {
		geminiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}
	if req.Stop != nil && len(req.Stop) > 0 {
		geminiReq.GenerationConfig.StopSequences = req.Stop
	}

	messages := req.Messages
	// Find and set system instruction if it exists
	if len(messages) > 0 && messages[0].Role == "system" {
		geminiReq.SystemInstruction = &GeminiContent{
			// Role for system instructions is not explicitly set, "user" is a placeholder
			Role:  "user",
			Parts: p.contentToGeminiParts(messages[0].Content),
		}
		// Remove system message from the list
		messages = messages[1:]
	}

	contents := make([]GeminiContent, 0, len(messages))
	for _, msg := range messages {
		role := "user" // Default role
		if msg.Role == "assistant" {
			role = "model"
		} else if msg.Role == "system" {
			// If a system message is found not at the beginning, ignore it for contents.
			// The Gemini API expects only one system instruction at the start.
			logger.Warn(ctx, logTag, "System message found in a position other than the first message, it will be ignored.")
			continue
		}
		contents = append(contents, GeminiContent{
			Role:  role,
			Parts: p.contentToGeminiParts(msg.Content),
		})
	}
	geminiReq.Contents = contents

	return geminiReq, nil
}

// contentToGeminiParts converts llm.Message.Content to []GeminiPart
// Supports:
// - string: simple text content
// - []interface{}: multimodal content array with format:
//   - {"type": "text", "text": "..."}
//   - {"type": "image_url", "image_url": {"url": "data:<mime>;base64,<data>"}}
func (p *Provider) contentToGeminiParts(content interface{}) []GeminiPart {
	// Case 1: simple string content
	if s, ok := content.(string); ok {
		return []GeminiPart{{Text: s}}
	}

	// Case 2: multimodal content array
	arr, ok := content.([]interface{})
	if !ok {
		// Fallback: convert to string
		return []GeminiPart{{Text: llm.ContentString(content)}}
	}

	parts := make([]GeminiPart, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		itemType, _ := m["type"].(string)
		switch itemType {
		case "text":
			if text, ok := m["text"].(string); ok {
				parts = append(parts, GeminiPart{Text: text})
			}
		case "image_url":
			// Parse {"image_url": {"url": "data:image/png;base64,..."}}
			imageURL, ok := m["image_url"].(map[string]interface{})
			if !ok {
				continue
			}
			url, ok := imageURL["url"].(string)
			if !ok {
				continue
			}
			// Parse data URI: data:<mime>;base64,<data>
			if inlineData := parseDataURI(url); inlineData != nil {
				parts = append(parts, GeminiPart{InlineData: inlineData})
			}
		}
	}

	if len(parts) == 0 {
		// Fallback if no valid parts parsed
		return []GeminiPart{{Text: llm.ContentString(content)}}
	}
	return parts
}

// parseDataURI parses a data URI (data:<mime>;base64,<data>) into InlineData
func parseDataURI(uri string) *InlineData {
	const prefix = "data:"
	if !strings.HasPrefix(uri, prefix) {
		return nil
	}
	// Format: data:<mime>;base64,<data>
	rest := uri[len(prefix):]
	semicolonIdx := strings.Index(rest, ";base64,")
	if semicolonIdx == -1 {
		return nil
	}
	mimeType := rest[:semicolonIdx]
	data := rest[semicolonIdx+len(";base64,"):]
	return &InlineData{
		MimeType: mimeType,
		Data:     data,
	}
}

func (p *Provider) toLLMChatResponse(geminiResp *GeminiChatResponse) *llm.ChatResponse {
	var content string
	var finishReason string
	if len(geminiResp.Candidates) > 0 {
		// Assuming the first candidate is the one we want
		candidate := geminiResp.Candidates[0]
		if len(candidate.Content.Parts) > 0 {
			content = candidate.Content.Parts[0].Text
		}
		finishReason = candidate.FinishReason
	}
	// Note: Gemini API v1beta for non-streaming chat doesn't return token usage.
	// This would need to be fetched from a different field or API if available.
	return &llm.ChatResponse{
		Content:      content,
		FinishReason: finishReason,
	}
}

// Ensure *Provider implements the interfaces
var _ llm.Chatter = (*Provider)(nil)
var _ llm.Embedder = (*Provider)(nil)
