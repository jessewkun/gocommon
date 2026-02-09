// Package gemini an provider for google gemini
package gemini

import (
	"fmt"
	"time"

	xhttp "github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/llm"
	"github.com/spf13/cast"
)

const providerName = "gemini"

func init() {
	llm.Register(providerName, func(config interface{}) (llm.Provider, error) {
		cfg, err := parseConfig(config)
		if err != nil {
			return nil, fmt.Errorf("parsing gemini config: %w", err)
		}
		// The client timeout will be overridden by the per-request timeout from cfg.Timeout
		client := xhttp.NewClient(xhttp.Option{})
		return NewProvider(client, cfg), nil
	})
}

// parseConfig supports Config or map for flexibility
func parseConfig(config interface{}) (Config, error) {
	if c, ok := config.(Config); ok {
		if c.APIURL == "" {
			c.APIURL = "https://generativelanguage.googleapis.com"
		}
		return c, nil
	}
	m, ok := config.(map[string]interface{})
	if !ok {
		return Config{}, fmt.Errorf("config must be of type gemini.Config or map[string]interface{}")
	}
	cfg := Config{
		APIKey: cast.ToString(m["api_key"]),
		APIURL: cast.ToString(m["api_url"]),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("api_key is required for gemini provider")
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "https://generativelanguage.googleapis.com"
	}
	if v, has := m["timeout"]; has {
		if d, ok := v.(time.Duration); ok {
			cfg.Timeout = d
		}
	}
	return cfg, nil
}
