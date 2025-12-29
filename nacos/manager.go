// Package nacos 提供 Nacos 连接管理功能
package nacos

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jessewkun/gocommon/logger"
)

// Manager 用于统一管理 Nacos 客户端实例
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewManager 创建和初始化一个 Manager 实例
func NewManager(configs map[string]*Config) (*Manager, error) {
	mgr := &Manager{
		clients: make(map[string]*Client),
	}
	var allErrors []error

	for clientName, conf := range configs {
		if err := setDefaultConfig(conf); err != nil {
			e := fmt.Errorf("nacos %s setDefaultConfig error: %w", clientName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue
		}
		client, err := newClient(conf)
		if err != nil {
			e := fmt.Errorf("connect to nacos %s failed, error: %w", clientName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			// 即使连接失败，也创建一个空的 client 占位，以便后续 HealthCheck 能发现
		}
		mgr.clients[clientName] = client
		if err == nil {
			logger.Info(context.Background(), TAG, "create nacos client %s succ, host: %s", clientName, conf.Host)
		}
	}

	if len(allErrors) > 0 {
		// Go 1.20+
		return mgr, errors.Join(allErrors...)
	}
	return mgr, nil
}

// GetClient 获取指定名称的 Nacos 客户端
func (m *Manager) GetClient(name string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, ok := m.clients[name]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("nacos client '%s' not found", name)
}

// Close 关闭所有 Nacos 客户端连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, client := range m.clients {
		if client != nil {
			if err := client.Close(); err != nil {
				e := fmt.Errorf("close nacos client %s failed: %w", name, err)
				errs = append(errs, e)
				logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			} else {
				logger.Info(context.Background(), TAG, "close nacos client %s succ", name)
			}
		}
	}

	// 清空 map
	m.clients = make(map[string]*Client)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
