package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
)

// NewClient 创建新的 ES 客户端
func NewClient(cfg Config) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}
	return &Client{ES: es}, nil
}
