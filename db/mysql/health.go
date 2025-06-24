package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"gorm.io/gorm"
)

// HealthCheck mysql健康检查
func HealthCheck() map[string]*HealthStatus {
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先获取所有连接信息，避免在Ping时持有锁
	connList.mu.RLock()
	connections := make(map[string]*gorm.DB)
	for dbName, db := range connList.conns {
		connections[dbName] = db
	}
	connList.mu.RUnlock()

	// 执行健康检查
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

		// 执行Ping检查
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
