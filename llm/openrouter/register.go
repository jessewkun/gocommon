// Package openrouter 实现 OpenRouter/OpenAI 兼容的 Chat API，并在 init 中注册到 llm，业务通过 llm.NewProvider("openrouter", config) 获取。
package openrouter

import (
	"fmt"
	"time"

	xhttp "github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/llm"
	"github.com/spf13/cast"
)

func init() {
	llm.Register(providerName, func(config interface{}) (llm.Provider, error) {
		cfg, err := parseConfig(config)
		if err != nil {
			return nil, err
		}
		// 注意：这里的 Timeout 最终会被 provider.go 中传递的 p.cfg.Timeout 覆盖
		// 仅作为没有在请求中指定 Timeout 时的 http client 默认值
		client := xhttp.NewClient(xhttp.Option{
			Timeout: cfg.Timeout,
			Headers: map[string]string{"Content-Type": "application/json"},
		})
		return NewProvider(client, cfg), nil
	})
}

// parseConfig 支持 Config 或 map（业务无需引用 openrouter 即可用 map 配置）
func parseConfig(config interface{}) (Config, error) {
	if c, ok := config.(Config); ok {
		if c.APIKey == "" {
			return Config{}, fmt.Errorf("openrouter: api_key is required")
		}
		if c.APIURL == "" {
			c.APIURL = "https://openrouter.ai/api/v1"
		}
		return c, nil
	}
	m, ok := config.(map[string]interface{})
	if !ok {
		return Config{}, fmt.Errorf("openrouter: config 须为 openrouter.Config 或 map[string]interface{}")
	}
	apiURL := cast.ToString(m["api_url"])
	if apiURL == "" {
		apiURL = cast.ToString(m["base_url"])
	}
	cfg := Config{
		APIURL: apiURL,
		APIKey: cast.ToString(m["api_key"]),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("openrouter: api_key is required")
	}
	if v, has := m["timeout"]; has {
		if d, ok := v.(time.Duration); ok {
			cfg.Timeout = d
		}
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "https://openrouter.ai/api/v1"
	}
	return cfg, nil
}
