package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Manager 用于统一管理 MongoDB 连接状态
type Manager struct {
	conns   map[string]*mongo.Client
	configs map[string]*Config
	mu      sync.RWMutex
}

// NewManager 创建和初始化一个 Manager 实例
func NewManager(configs map[string]*Config) (*Manager, error) {
	mgr := &Manager{
		conns:   make(map[string]*mongo.Client),
		configs: configs,
	}
	var allErrors []error

	for dbName, conf := range configs {
		if err := setMongoDefaultConfig(conf); err != nil {
			e := fmt.Errorf("mongodb %s setDefaultConfig error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue // 继续尝试下一个
		}

		client, err := newClient(conf)
		if err != nil {
			e := fmt.Errorf("connect to mongodb %s failed, error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue // 继续尝试下一个
		}
		mgr.conns[dbName] = client
		logger.Info(context.Background(), TAG, "connect to mongodb %s succ", dbName)
	}

	if len(allErrors) > 0 {
		return mgr, errors.Join(allErrors...)
	}
	return mgr, nil
}

// GetConn 获取 MongoDB 客户端
func (m *Manager) GetConn(dbIns string) (*mongo.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, ok := m.conns[dbIns]; ok && client != nil {
		return client, nil
	}
	return nil, fmt.Errorf("mongodb instance '%s' not found or is not connected", dbIns)
}

// Close 关闭所有 MongoDB 连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for dbName, client := range m.conns {
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := client.Disconnect(ctx); err != nil {
				e := fmt.Errorf("disconnect mongodb %s failed: %w", dbName, err)
				errs = append(errs, e)
				logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			} else {
				logger.Info(context.Background(), TAG, "disconnect mongodb %s succ", dbName)
			}
			cancel()
		}
	}

	// 清空连接列表
	m.conns = make(map[string]*mongo.Client)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// HealthCheck 执行 MongoDB 健康检查
func (m *Manager) HealthCheck() map[string]*HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.mu.RLock()
	connections := make(map[string]*mongo.Client, len(m.conns))
	for dbName, client := range m.conns {
		connections[dbName] = client
	}
	m.mu.RUnlock()

	resp := make(map[string]*HealthStatus)
	for dbName, client := range connections {
		status := &HealthStatus{
			Timestamp: time.Now().UnixMilli(),
		}
		conf := m.configs[dbName]

		if client == nil {
			status.Status = "error"
			status.Error = "client is nil"
			resp[dbName] = status
			continue
		}

		// 获取连接池状态
		status.InUse = client.NumberSessionsInProgress()
		if conf != nil {
			status.MaxPool = conf.MaxPoolSize
		}
		status.Available = status.MaxPool - status.InUse
		status.Idle = 0 // MongoDB 驱动不直接提供空闲连接数

		// 执行Ping检查
		startTime := time.Now()
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			status.Status = "error"
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				status.Error = "timeout"
			case errors.Is(err, mongo.ErrClientDisconnected):
				status.Error = "client disconnected"
			default:
				status.Error = err.Error()
			}
			logger.ErrorWithMsg(ctx, TAG, "ping mongodb %s failed, error: %s", dbName, err)
		} else {
			status.Status = "success"
		}
		status.Latency = time.Since(startTime).Milliseconds()
		resp[dbName] = status
	}
	return resp
}
