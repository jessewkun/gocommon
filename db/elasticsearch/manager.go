// package elasticsearch provides elasticsearch client management
package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jessewkun/gocommon/logger"
)

// Manager 用于统一管理 Elasticsearch 连接
type Manager struct {
	conns map[string]*Client
	mu    sync.RWMutex
}

// NewManager 创建并初始化一个 Manager 实例
func NewManager(configs map[string]*Config) (*Manager, error) {
	mgr := &Manager{
		conns: make(map[string]*Client),
	}
	var allErrors []error

	for dbName, conf := range configs {
		esCfg := elasticsearch.Config{
			Addresses: conf.Addresses,
			Username:  conf.Username,
			Password:  conf.Password,
		}
		if conf.IsLog {
			slowThreshold := time.Duration(conf.SlowThreshold) * time.Millisecond
			if slowThreshold == 0 {
				slowThreshold = 200 * time.Millisecond
			}
			esCfg.Transport = newLoggingTransport(slowThreshold)
		}

		es, err := elasticsearch.NewClient(esCfg)
		if err != nil {
			e := fmt.Errorf("create elasticsearch client %s failed: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue // 创建失败则不加入连接池
		}

		// 验证连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := es.Info(es.Info.WithContext(ctx))
		cancel()
		if err != nil || (res != nil && res.IsError()) {
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			} else if res != nil {
				errMsg = res.String()
			}
			e := fmt.Errorf("ping elasticsearch %s failed: %s", dbName, errMsg)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			if res != nil {
				res.Body.Close()
			}
			continue
		}
		if res != nil {
			res.Body.Close()
		}

		mgr.conns[dbName] = &Client{ES: es}
		logger.Info(context.Background(), TAG, "connect to elasticsearch %s succ", dbName)
	}

	if len(allErrors) > 0 {
		return mgr, errors.Join(allErrors...)
	}
	return mgr, nil
}

// GetConn 获取 ES 连接
func (m *Manager) GetConn(dbName string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, ok := m.conns[dbName]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("elasticsearch client '%s' not found", dbName)
}

// Close 关闭所有 ES 连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// go-elasticsearch client 不需要显式关闭
	logger.Info(context.Background(), TAG, "clearing all elasticsearch connections")
	m.conns = make(map[string]*Client)
	return nil
}

// HealthCheck 执行 Elasticsearch 健康检查（并发版）
func (m *Manager) HealthCheck() map[string]*HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.mu.RLock()
	connections := make(map[string]*Client, len(m.conns))
	for dbName, db := range m.conns {
		connections[dbName] = db
	}
	m.mu.RUnlock()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		resp = make(map[string]*HealthStatus)
	)

	for dbName, client := range connections {
		wg.Add(1)
		go func(name string, c *Client) {
			defer wg.Done()

			status := &HealthStatus{
				Timestamp: time.Now().UnixMilli(),
			}

			start := time.Now()
			res, err := c.ES.Cluster.Health(
				c.ES.Cluster.Health.WithContext(ctx),
			)
			status.Latency = time.Since(start).Milliseconds()

			if err != nil {
				status.Status = "error"
				status.Error = err.Error()
				if res != nil {
					// 确保在错误情况下也关闭 body
					res.Body.Close()
				}
			} else {
				if res.IsError() {
					status.Status = "error"
					status.Error = res.Status()
				} else {
					status.Status = "success"
				}
				// 显式关闭 body，替代 for 循环中的 defer
				res.Body.Close()
			}

			mu.Lock()
			resp[name] = status
			mu.Unlock()
		}(dbName, client)
	}

	wg.Wait()
	return resp
}
