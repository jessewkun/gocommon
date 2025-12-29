// Package alarm 提供统一报警接口，支持同时发送到多个配置的渠道
package alarm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
)

// Sender 是一个适配器，用于实现 logger.Alerter 接口
type Sender struct{}

// Send 方法实现了 logger.Alerter 接口
func (s *Sender) Send(ctx context.Context, title string, content []string) error {
	return SendAlarm(ctx, title, content)
}

// SendAlarm 统一报警接口，支持同时发送到多个配置的渠道
func SendAlarm(ctx context.Context, title string, content []string) error {
	if Cfg == nil {
		return fmt.Errorf("alarm config is not initialized")
	}

	// 检查是否有任何渠道被配置，如果没有，则提前返回
	barkConfigured := Cfg.Bark != nil && len(Cfg.Bark.BarkIds) > 0
	feishuConfigured := Cfg.Feishu != nil && len(Cfg.Feishu.WebhookURL) > 0
	if !barkConfigured && !feishuConfigured {
		return fmt.Errorf("no alarm channels configured")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []string

	// 发送到 Bark 渠道
	if barkConfigured {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Cfg.Bark.Send(ctx, title, content); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("bark failed: %v", err))
				mu.Unlock()
				log.Printf("Failed to send alarm via Bark: %v", err)
			}
		}()
	}

	// 发送到 Feishu 渠道
	if feishuConfigured {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Cfg.Feishu.Send(ctx, title, content); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("feishu failed: %v", err))
				mu.Unlock()
				log.Printf("Failed to send alarm via Feishu: %v", err)
			}
		}()
	}

	wg.Wait()

	// 如果有任何错误，合并所有错误信息并返回
	if len(errs) > 0 {
		return fmt.Errorf("failed to send alarm to some channels: [%s]", strings.Join(errs, "; "))
	}

	return nil
}
