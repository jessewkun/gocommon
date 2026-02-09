// Package gemini an provider for google gemini
package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	xhttp "github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestProvider(serverURL string) *Provider {
	cfg := Config{
		APIKey:  "test-key",
		APIURL:  serverURL,
		Timeout: 5 * time.Second,
	}
	client := xhttp.NewClient(xhttp.Option{})
	return NewProvider(client, cfg)
}

func TestGeminiProvider_Chat_WithSystemPrompt(t *testing.T) {
	// 1. Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, ":generateContent"))

		// Check request body
		var reqBody GeminiChatRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Assert system prompt is handled correctly
		require.NotNil(t, reqBody.SystemInstruction)
		assert.Equal(t, "You are a helpful assistant.", reqBody.SystemInstruction.Parts[0].Text)
		assert.Equal(t, "user", reqBody.SystemInstruction.Role) // Role is a placeholder, as expected

		// Assert user message is present
		require.Len(t, reqBody.Contents, 1)
		assert.Equal(t, "user", reqBody.Contents[0].Role)
		assert.Equal(t, "Hello", reqBody.Contents[0].Parts[0].Text)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		resp := GeminiChatResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{{Text: "Hi there!"}},
						Role:  "model",
					},
					FinishReason: "STOP",
				},
			},
		}
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	// 2. Create provider and request
	provider := newTestProvider(server.URL)
	req := &llm.ChatRequest{
		Model: "gemini-pro",
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
	}

	// 3. Call Chat and assert
	resp, err := provider.Chat(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Hi there!", resp.Content)
	assert.Equal(t, "STOP", resp.FinishReason)
}

func TestGeminiProvider_ChatStream(t *testing.T) {
	// 1. Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, ":streamGenerateContent"))

		w.Header().Set("Content-Type", "text/event-stream")
		// Send a few chunks
		_, _ = w.Write([]byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"Hello\"}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \" \"}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"World\"}]}}]}\n\n"))
		// Send final chunk with usage metadata
		_, _ = w.Write([]byte("data: {\"candidates\": [{\"finishReason\": \"STOP\"}], \"usageMetadata\": {\"promptTokenCount\": 10, \"totalTokenCount\": 25}}\n\n"))
	}))
	defer server.Close()

	// 2. Create provider and request
	provider := newTestProvider(server.URL)
	req := &llm.ChatRequest{
		Model:    "gemini-pro-vision",
		Messages: []llm.Message{{Role: "user", Content: "Say Hello World"}},
	}

	// 3. Call ChatStream and assert
	var chunks []string
	callback := func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}

	fullResp, err := provider.ChatStream(context.Background(), req, callback)
	require.NoError(t, err)
	require.NotNil(t, fullResp)

	// Assert chunks and final response
	assert.Equal(t, []string{"Hello", " ", "World"}, chunks)
	assert.Equal(t, "Hello World", fullResp.Content)
	assert.Equal(t, "STOP", fullResp.FinishReason)
	assert.Equal(t, 10, fullResp.Usage.PromptTokens)
	assert.Equal(t, 25, fullResp.Usage.TotalTokens)
	assert.Equal(t, 15, fullResp.Usage.CompletionTokens)
}

func TestGeminiProvider_Chat_WithMultimodal(t *testing.T) {
	// 1. Setup mock server that validates multimodal request format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, ":generateContent"))

		// Check request body
		var reqBody GeminiChatRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Assert user message contains both text and image parts
		require.Len(t, reqBody.Contents, 1)
		content := reqBody.Contents[0]
		assert.Equal(t, "user", content.Role)
		require.Len(t, content.Parts, 2)

		// First part should be text
		assert.Equal(t, "What is in this image?", content.Parts[0].Text)
		assert.Nil(t, content.Parts[0].InlineData)

		// Second part should be inline image data
		assert.Empty(t, content.Parts[1].Text)
		require.NotNil(t, content.Parts[1].InlineData)
		assert.Equal(t, "image/png", content.Parts[1].InlineData.MimeType)
		assert.Equal(t, "iVBORw0KGgo=", content.Parts[1].InlineData.Data)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		resp := GeminiChatResponse{
			Candidates: []GeminiCandidate{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{{Text: "I see a test image."}},
						Role:  "model",
					},
					FinishReason: "STOP",
				},
			},
		}
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	// 2. Create provider and multimodal request
	provider := newTestProvider(server.URL)
	req := &llm.ChatRequest{
		Model: "gemini-pro-vision",
		Messages: []llm.Message{
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{"type": "text", "text": "What is in this image?"},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": "data:image/png;base64,iVBORw0KGgo=",
						},
					},
				},
			},
		},
	}

	// 3. Call Chat and assert
	resp, err := provider.Chat(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "I see a test image.", resp.Content)
	assert.Equal(t, "STOP", resp.FinishReason)
}
