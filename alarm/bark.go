package alarm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	TAGNAME = "ALARM"
	barkAPI = "https://api.day.app/%s/%s/%s"
)

var (
	barkIds []string
	mu      sync.RWMutex
	client  *http.Client
)

const MaxRetry = 3

// InitBark 初始化 Bark 报警
func InitBark(config *Config) error {
	// 如果 BarkIds 为空，则不初始化
	if len(config.BarkIds) == 0 {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	barkIds = config.BarkIds
	client = &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return nil
}

// SendBark 发送 Bark 报警
func SendBark(ctx context.Context, title, content string) error {
	mu.RLock()
	ids := barkIds
	mu.RUnlock()

	if len(ids) == 0 {
		return fmt.Errorf("bark ids not initialized")
	}

	for _, id := range ids {
		if err := sendWithRetry(ctx, id, title, content); err != nil {
			log.Printf("Failed to send bark to %s: %v", id, err)
			continue
		}
		log.Printf("Successfully sent bark to %s", id)
	}

	return nil
}

// sendWithRetry 带重试的发送
func sendWithRetry(ctx context.Context, id, title, content string) error {
	var lastErr error
	for i := 0; i < MaxRetry; i++ {
		if err := send(ctx, id, title, content); err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return nil
	}
	return lastErr
}

// send 发送单个 Bark 消息
func send(ctx context.Context, id, title, content string) error {
	url := fmt.Sprintf(barkAPI, id, url.QueryEscape(title), url.QueryEscape(content))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
