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
	client *http.Client
	mu     sync.RWMutex
)

const MaxRetry = 3

func Init() error {
	mu.Lock()
	defer mu.Unlock()

	if Cfg == nil || len(Cfg.BarkIds) == 0 {
		client = nil
		return nil
	}

	client = &http.Client{
		Timeout: time.Duration(Cfg.Timeout) * time.Second,
	}

	return nil
}

func SendBark(ctx context.Context, title, content string) error {
	mu.RLock()
	ids := Cfg.BarkIds
	c := client
	mu.RUnlock()

	if c == nil || len(ids) == 0 {
		return fmt.Errorf("bark is not configured or initialized")
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

func send(ctx context.Context, id, title, content string) error {
	url := fmt.Sprintf(barkAPI, id, url.QueryEscape(title), url.QueryEscape(content))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	mu.RLock()
	c := client
	mu.RUnlock()

	if c == nil {
		return fmt.Errorf("http client for bark not initialized")
	}

	resp, err := c.Do(req)
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
