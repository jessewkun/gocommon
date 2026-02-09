// Package openrouter an provider for openrouter
package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestOpenRouterProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		resp := OpenAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "openai/gpt-3.5-turbo",
			Choices: []OpenAIChatChoice{
				{
					Index: 0,
					Message: llm.Message{
						Role:    "assistant",
						Content: "Hello from OpenRouter!",
					},
					FinishReason: "stop",
				},
			},
			Usage: llm.Usage{
				PromptTokens:     10,
				CompletionTokens: 10,
				TotalTokens:      20,
			},
		}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL)
	req := &llm.ChatRequest{
		Model:    "openai/gpt-3.5-turbo",
		Messages: []llm.Message{{Role: "user", Content: "Hello"}},
	}

	resp, err := provider.Chat(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Hello from OpenRouter!", resp.Content)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Equal(t, 20, resp.Usage.TotalTokens)
}

func TestOpenRouterProvider_ChatStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"id\":\"1\",\"object\":\"completion.chunk\",\"created\":1702263749,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hello\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"id\":\"1\",\"object\":\"completion.chunk\",\"created\":1702263749,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" World\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"id\":\"1\",\"object\":\"completion.chunk\",\"created\":1702263749,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	provider := newTestProvider(server.URL)
	req := &llm.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []llm.Message{{Role: "user", Content: "Say Hello World"}},
	}

	var chunks []string
	fullResp, err := provider.ChatStream(context.Background(), req, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, fullResp)
	assert.Equal(t, []string{"Hello", " World"}, chunks)
	assert.Equal(t, "Hello World", fullResp.Content)
	assert.Equal(t, "stop", fullResp.FinishReason)
}

func TestOpenRouterProvider_CreateEmbeddings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/embeddings", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		resp := OpenAIEmbeddingResponse{
			Object: "list",
			Data: []OpenAIEmbeddingData{
				{
					Object:    "embedding",
					Embedding: []float32{0.1, 0.2, 0.3},
					Index:     0,
				},
			},
			Model: "text-embedding-ada-002",
			Usage: llm.Usage{PromptTokens: 8, TotalTokens: 8},
		}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	provider := newTestProvider(server.URL)
	req := &llm.EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{"The quick brown fox jumps over the lazy dog"},
	}

	resp, err := provider.CreateEmbeddings(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "text-embedding-ada-002", resp.Model)
	assert.Equal(t, 8, resp.Usage.TotalTokens)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, resp.Data[0].Vector)
}
