// Package alarm 提供统一报警接口，支持同时发送到多个配置的渠道
package alarm

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// SendAlarm 统一报警接口，支持同时发送到多个配置的渠道
func SendAlarm(ctx context.Context, title string, content []string) error {
	if Cfg == nil {
		return fmt.Errorf("alarm config is not initialized")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// 发送到 Bark 渠道
	if Cfg.Bark != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Cfg.Bark.Send(ctx, title, content); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("bark failed: %v", err))
				mu.Unlock()
				log.Printf("Failed to send alarm via Bark: %v", err)
			}
		}()
	}

	// 发送到 Feishu 渠道
	if Cfg.Feishu != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Cfg.Feishu.Send(ctx, title, content); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("feishu failed: %v", err))
				mu.Unlock()
				log.Printf("Failed to send alarm via Feishu: %v", err)
			}
		}()
	}

	wg.Wait()

	// 如果没有配置任何渠道，返回错误
	if Cfg.Bark == nil || len(Cfg.Bark.BarkIds) == 0 {
		if Cfg.Feishu == nil || Cfg.Feishu.WebhookURL == "" {
			return fmt.Errorf("no alarm channels configured")
		}
	}

	// 如果有任何错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
