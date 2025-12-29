// Package mysql 提供 MySQL 数据库连接管理功能
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"gorm.io/gorm"
)

type Manager struct {
	conns map[string]*gorm.DB
	mu    sync.RWMutex
}

// NewManager 创建和初始化一个 Manager 实例
func NewManager(configs map[string]*Config) (*Manager, error) {
	mgr := &Manager{
		conns: make(map[string]*gorm.DB),
	}
	var allErrors []error

	for dbName, conf := range configs {
		if err := setDefaultConfig(conf); err != nil {
			e := fmt.Errorf("mysql %s setDefaultConfig error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			continue
		}
		db, err := newClient(conf)
		if err != nil {
			e := fmt.Errorf("connect to mysql %s failed, error: %w", dbName, err)
			allErrors = append(allErrors, e)
			logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			// 即使连接失败，也保存配置信息，以便后续重试或检查
			// 使用 nil 表示连接失败，但配置已处理
			mgr.conns[dbName] = nil
			continue
		}
		mgr.conns[dbName] = db
		logger.Info(context.Background(), TAG, "connect to mysql %s succ", dbName)
	}

	return mgr, errors.Join(allErrors...)
}

// GetConn 获取数据库连接
func (m *Manager) GetConn(dbIns string) (*gorm.DB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if db, ok := m.conns[dbIns]; ok {
		if db == nil {
			return nil, fmt.Errorf("mysql conn '%s' connection failed, please check configuration", dbIns)
		}
		return db, nil
	}
	return nil, fmt.Errorf("mysql conn '%s' is not found", dbIns)
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for dbName, db := range m.conns {
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				e := fmt.Errorf("get sql.DB for mysql %s failed: %w", dbName, err)
				errs = append(errs, e)
				logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
				continue
			}

			if err := sqlDB.Close(); err != nil {
				e := fmt.Errorf("close mysql %s failed: %w", dbName, err)
				errs = append(errs, e)
				logger.ErrorWithMsg(context.Background(), TAG, "%s", e.Error())
			} else {
				logger.Info(context.Background(), TAG, "close mysql %s succ", dbName)
			}
		}
	}

	// 清空连接列表
	m.conns = make(map[string]*gorm.DB)

	return errors.Join(errs...)
}

// HealthCheck 执行mysql健康检查
func (m *Manager) HealthCheck() map[string]*HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先获取所有连接信息，避免在Ping时持有锁
	m.mu.RLock()
	connections := make(map[string]*gorm.DB, len(m.conns))
	for dbName, db := range m.conns {
		connections[dbName] = db
	}
	m.mu.RUnlock()

	resp := make(map[string]*HealthStatus)
	for dbName, db := range connections {
		status := &HealthStatus{
			Timestamp: time.Now().UnixMilli(),
		}

		sqldb, err := db.DB()
		if err != nil {
			status.Status = "error"
			status.Error = fmt.Sprintf("get db instance error: %v", err)
			logger.ErrorWithMsg(ctx, TAG, "get mysql %s db failed, error: %s", dbName, err)
			resp[dbName] = status
			continue
		}

		// 获取连接池状态
		stats := sqldb.Stats()
		status.MaxOpen = stats.MaxOpenConnections
		status.Open = stats.OpenConnections
		status.InUse = stats.InUse
		status.Idle = stats.Idle
		status.WaitCount = stats.WaitCount
		status.WaitTime = stats.WaitDuration.Nanoseconds()

		startTime := time.Now()
		if err := sqldb.PingContext(ctx); err != nil {
			status.Status = "error"
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				status.Error = "timeout"
			case errors.Is(err, sql.ErrConnDone):
				status.Error = "connection closed"
			default:
				status.Error = err.Error()
			}
			logger.ErrorWithMsg(ctx, TAG, "ping mysql %s failed, error: %s", dbName, err)
		} else {
			status.Status = "success"
		}
		status.Latency = time.Since(startTime).Milliseconds()

		resp[dbName] = status
	}

	return resp
}
