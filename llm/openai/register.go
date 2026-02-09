// Package openai an provider for openai
package openai

import (
	"fmt"
	"time"

	xhttp "github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/llm"
	"github.com/spf13/cast"
)

const providerName = "openai"

func init() {
	llm.Register(providerName, func(config interface{}) (llm.Provider, error) {
		cfg, err := parseConfig(config)
		if err != nil {
			return nil, fmt.Errorf("parsing openai config: %w", err)
		}
		// The client timeout will be overridden by the per-request timeout from cfg.Timeout
		client := xhttp.NewClient(xhttp.Option{})
		return NewProvider(client, cfg), nil
	})
}

// parseConfig supports Config or map for flexibility
func parseConfig(config interface{}) (Config, error) {
	if c, ok := config.(Config); ok {
		// Set default API URL if not provided
		if c.APIURL == "" {
			c.APIURL = "https://api.openai.com/v1"
		}
		return c, nil
	}
	m, ok := config.(map[string]interface{})
	if !ok {
		return Config{}, fmt.Errorf("config must be of type openai.Config or map[string]interface{}")
	}
	cfg := Config{
		APIKey: cast.ToString(m["api_key"]),
		APIURL: cast.ToString(m["api_url"]),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("api_key is required for openai provider")
	}
	if cfg.APIURL == "" {
		// Provide a default API URL if not specified
		cfg.APIURL = "https://api.openai.com/v1"
	}
	// Add timeout parsing logic
	if v, has := m["timeout"]; has {
		if d, ok := v.(time.Duration); ok {
			cfg.Timeout = d
		}
	}
	return cfg, nil
}
