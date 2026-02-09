package openrouter

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

const providerName = "openrouter"
const logTag = "LLM_OPENROUTER"

// Provider implements llm.Chatter and llm.Embedder
type Provider struct {
	client *xhttp.Client
	cfg    Config
}

// NewProvider creates a new OpenRouter Provider
func NewProvider(client *xhttp.Client, cfg Config) *Provider {
	return &Provider{client: client, cfg: cfg}
}

// Name implements llm.Provider
func (p *Provider) Name() string { return providerName }

// Chat implements llm.Chatter for non-streaming chat
func (p *Provider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	body, err := p.buildChatBody(req, false)
	if err != nil {
		return nil, fmt.Errorf("%s: building chat request: %w", providerName, err)
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%s: marshalling chat request: %w", providerName, err)
	}

	apiURL := p.cfg.APIURL + "/chat/completions"
	headers := p.buildAuthHeaders()

	resp, err := p.client.Post(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: headers,
		Timeout: p.cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: chat api call failed: %w", providerName, err)
	}
	if resp == nil || len(resp.Body) == 0 {
		return nil, fmt.Errorf("%s: empty response", providerName)
	}

	var apiResp OpenAIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("%s: unmarshalling chat response: %w (body: %s)", providerName, err, string(resp.Body))
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: api error: %s", providerName, apiResp.Error.Message)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("%s: no choices returned", providerName)
	}

	contentStr := ""
	if c, ok := apiResp.Choices[0].Message.Content.(string); ok {
		contentStr = c
	}
	return &llm.ChatResponse{
		Content:      strings.TrimSpace(contentStr),
		FinishReason: apiResp.Choices[0].FinishReason,
		Usage:        apiResp.Usage,
		RawResponse:  resp.Body,
	}, nil
}

// ChatStream implements llm.Chatter for streaming chat
func (p *Provider) ChatStream(ctx context.Context, req *llm.ChatRequest, callback func(chunk string) error) (*llm.ChatResponse, error) {
	body, err := p.buildChatBody(req, true)
	if err != nil {
		return nil, fmt.Errorf("%s: building stream request: %w", providerName, err)
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%s: marshalling stream request: %w", providerName, err)
	}

	apiURL := p.cfg.APIURL + "/chat/completions"
	headers := p.buildAuthHeaders()
	headers["Accept"] = "text/event-stream"

	var fullContent strings.Builder
	var finalResponse *llm.ChatResponse

	timeout := p.cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	err = p.client.PostStream(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: headers,
		Timeout: timeout,
	}, func(line []byte) error {
		if !bytes.HasPrefix(line, []byte("data: ")) {
			return nil
		}
		dataStr := bytes.TrimPrefix(line, []byte("data: "))
		if string(dataStr) == "[DONE]" {
			return nil
		}

		var streamResp OpenAIStreamResponse
		if err := json.Unmarshal(dataStr, &streamResp); err != nil {
			logger.WarnWithField(ctx, logTag, "failed to parse stream line", map[string]interface{}{"line": string(line), "error": err.Error()})
			return nil
		}

		if streamResp.Error != nil {
			return fmt.Errorf("%s: stream error: %s", providerName, streamResp.Error.Message)
		}

		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]
			if choice.Delta.Content != "" {
				chunk := choice.Delta.Content
				fullContent.WriteString(chunk)
				if err := callback(chunk); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
			if choice.FinishReason != "" {
				if finalResponse == nil {
					finalResponse = &llm.ChatResponse{}
				}
				finalResponse.FinishReason = choice.FinishReason
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: stream failed: %w", providerName, err)
	}

	if finalResponse == nil {
		finalResponse = &llm.ChatResponse{FinishReason: "unknown"}
	}
	finalResponse.Content = fullContent.String()
	return finalResponse, nil
}

// CreateEmbeddings implements llm.Embedder
func (p *Provider) CreateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	body := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openrouter: marshalling embedding request: %w", err)
	}

	apiURL := p.cfg.APIURL + "/embeddings"
	headers := p.buildAuthHeaders()

	resp, err := p.client.Post(ctx, xhttp.RequestPost{
		URL:     apiURL,
		Payload: bodyBytes,
		Headers: headers,
		Timeout: p.cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("openrouter: embedding api call failed: %w", err)
	}

	var openAIResp OpenAIEmbeddingResponse
	if err := json.Unmarshal(resp.Body, &openAIResp); err != nil {
		return nil, fmt.Errorf("openrouter: unmarshalling embedding response: %w (body: %s)", err, string(resp.Body))
	}
	if openAIResp.Error != nil {
		return nil, fmt.Errorf("openrouter: embedding api error: %s", openAIResp.Error.Message)
	}

	embeddings := make([]llm.Embedding, len(openAIResp.Data))
	for i, d := range openAIResp.Data {
		embeddings[i] = llm.Embedding{
			Index:  d.Index,
			Vector: d.Embedding,
			Object: d.Object,
		}
	}

	return &llm.EmbeddingResponse{
		Data:        embeddings,
		Model:       openAIResp.Model,
		Usage:       openAIResp.Usage,
		RawResponse: resp.Body,
	}, nil
}

func (p *Provider) buildChatBody(req *llm.ChatRequest, stream bool) (map[string]interface{}, error) {
	if req == nil {
		return nil, fmt.Errorf("%s: req cannot be nil", providerName)
	}
	msgs := make([]map[string]interface{}, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = map[string]interface{}{"role": m.Role, "content": m.Content}
	}
	body := map[string]interface{}{
		"model":       req.Model,
		"messages":    msgs,
		"temperature": req.Temperature,
		"stream":      stream,
	}
	if req.ResponseFormat != "" {
		body["response_format"] = map[string]interface{}{"type": req.ResponseFormat}
	}
	return body, nil
}

func (p *Provider) buildAuthHeaders() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + p.cfg.APIKey,
	}
}

// Ensure *Provider implements the interfaces
var _ llm.Chatter = (*Provider)(nil)
var _ llm.Embedder = (*Provider)(nil)
