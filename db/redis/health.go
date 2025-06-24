package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jessewkun/gocommon/logger"
)

// HealthCheck redis健康检查
func HealthCheck() map[string]map[string]*HealthStatus {
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先获取所有连接信息，避免在Ping时持有锁
	connList.mu.RLock()
	connections := make(map[string]map[string]*redis.Client)
	for dbName, conns := range connList.conns {
		connections[dbName] = make(map[string]*redis.Client)
		for addr, conn := range conns {
			connections[dbName][addr] = conn
		}
	}
	connList.mu.RUnlock()

	// 执行健康检查
	resp := make(map[string]map[string]*HealthStatus)
	for dbName, conns := range connections {
		resp[dbName] = make(map[string]*HealthStatus)
		for addr, conn := range conns {
			status := &HealthStatus{
				Timestamp: time.Now().UnixMilli(),
			}

			startTime := time.Now()
			_, err := conn.Ping(ctx).Result()
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
				logger.ErrorWithMsg(ctx, TAG, "redis ping db %s addr %s error %s", dbName, addr, err)
			} else {
				status.Status = "success"
			}

			resp[dbName][addr] = status
		}
	}

	return resp
}
