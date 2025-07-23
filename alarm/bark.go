package alarm

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

const (
	barkAPI = "https://api.day.app/%s/%s/%s"
)

type Bark struct {
	BarkIds []string `mapstructure:"bark_ids" json:"bark_ids"` // Bark 设备 ID 列表
}

// Send 发送 Bark 消息
func (b *Bark) Send(ctx context.Context, title string, content []string) error {
	if len(b.BarkIds) == 0 {
		return fmt.Errorf("bark is not configured")
	}

	// 将多行内容合并为单行，用换行符分隔
	contentStr := ""
	for i, line := range content {
		if i > 0 {
			contentStr += "\n"
		}
		contentStr += line
	}

	// 并发向所有 Bark 设备发送消息
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, barkID := range b.BarkIds {
		wg.Add(1)
		go func(barkID string) {
			defer wg.Done()

			// 构建URL
			url := fmt.Sprintf(barkAPI, barkID, url.QueryEscape(title), url.QueryEscape(contentStr))

			req := &HTTPRequest{
				Method: "GET",
				URL:    url,
			}

			if err := SendHTTPRequestWithRetry(ctx, req, MaxRetry); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("bark device %s failed: %v", barkID, err))
				mu.Unlock()
			}
		}(barkID)
	}

	wg.Wait()

	// 如果有任何错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
