package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoHealthCheck MongoDB 健康检查
func MongoHealthCheck() map[string]*MongoHealthStatus {
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先获取所有连接信息，避免在Ping时持有锁
	mongoConnList.mu.RLock()
	connections := make(map[string]*mongo.Client)
	for dbName, client := range mongoConnList.conns {
		connections[dbName] = client
	}
	mongoConnList.mu.RUnlock()

	// 执行健康检查
	resp := make(map[string]*MongoHealthStatus)
	for dbName, client := range connections {
		status := &MongoHealthStatus{
			Timestamp: time.Now().UnixMilli(),
		}

		if client == nil {
			status.Status = "error"
			status.Error = "client is nil"
			logger.ErrorWithMsg(ctx, TAGNAME, "mongodb %s client is nil", dbName)
			resp[dbName] = status
			continue
		}

		// 获取连接池状态
		poolStats := client.NumberSessionsInProgress()
		status.InUse = poolStats

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
			logger.ErrorWithMsg(ctx, TAGNAME, "ping mongodb %s failed, error: %s", dbName, err)
		} else {
			status.Status = "success"
		}
		status.Latency = time.Since(startTime).Milliseconds()

		// 尝试获取更详细的连接池信息
		// 注意：MongoDB Go 驱动不直接提供连接池统计信息，这里使用会话数量作为参考
		status.MaxPool = 100 // 默认值，实际值需要从配置中获取
		status.Idle = 0      // MongoDB 驱动不直接提供空闲连接数
		status.Available = status.MaxPool - status.InUse

		resp[dbName] = status
	}

	return resp
}
