// Package redis 提供 Redis 连接管理功能
package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jessewkun/gocommon/logger"
)

// Manager 用于统一管理 Redis 连接状态
type Manager struct {
	conns map[string]redis.UniversalClient
	mu    sync.RWMutex
}

// NewManager 创建和初始化一个 Manager 实例
func NewManager(configs map[string]*Config) (*Manager, error) {
	mgr := &Manager{
		conns: make(map[string]redis.UniversalClient),
	}
	var allErrors []error

	for dbName, conf := range configs {
		if err := setDefaultConfig(conf); err != nil {
			e := fmt.Errorf("redis %s setDefaultConfig error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue
		}
		client, err := newClient(conf)
		if err != nil {
			e := fmt.Errorf("connect to redis %s failed, error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			// 不再 continue，继续将 client 实例加入 map
		}
		mgr.conns[dbName] = client
		logger.Info(context.Background(), TAG, "create redis client %s succ, addrs: %v", dbName, conf.Addrs)
	}

	if len(allErrors) > 0 {
		return mgr, errors.Join(allErrors...)
	}
	return mgr, nil
}

// GetConn 获取 Redis 连接
func (m *Manager) GetConn(dbIns string) (redis.UniversalClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, ok := m.conns[dbIns]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("redis instance '%s' not found", dbIns)
}

// Close 关闭所有 Redis 连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for dbName, client := range m.conns {
		if client != nil {
			if err := client.Close(); err != nil {
				e := fmt.Errorf("close redis %s failed: %w", dbName, err)
				errs = append(errs, e)
				logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			} else {
				logger.Info(context.Background(), TAG, "close redis %s succ", dbName)
			}
		}
	}

	// 清空连接列表
	m.conns = make(map[string]redis.UniversalClient)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// HealthCheck 执行 Redis 健康检查
func (m *Manager) HealthCheck() map[string]*HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.mu.RLock()
	connections := make(map[string]redis.UniversalClient, len(m.conns))
	for dbName, db := range m.conns {
		connections[dbName] = db
	}
	m.mu.RUnlock()

	resp := make(map[string]*HealthStatus)
	for dbName, client := range connections {
		status := &HealthStatus{
			Timestamp: time.Now().UnixMilli(),
		}

		startTime := time.Now()
		_, err := client.Ping(ctx).Result()
		latency := time.Since(startTime).Milliseconds()
		status.Latency = latency

		if err != nil {
			status.Status = "error"
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				status.Error = "timeout"
			case errors.Is(err, redis.ErrClosed):
				status.Error = "connection closed"
			default:
				status.Error = err.Error()
			}
			logger.ErrorWithMsg(ctx, TAG, "redis ping db %s error %s", dbName, err)
		} else {
			status.Status = "success"
		}

		resp[dbName] = status
	}

	return resp
}
