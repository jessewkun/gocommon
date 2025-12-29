// Package elasticsearch 提供 Elasticsearch 连接管理功能
package elasticsearch

import (
	"errors"
	"time"
)

const TAG = "ELASTICSEARCH"

// Init 初始化 defaultManager
func Init() error {
	var err error
	defaultManager, err = NewManager(Cfgs)
	return err
}

// GetConn 获取指定名称的 ES 客户端
func GetConn(dbName string) (*Client, error) {
	if defaultManager == nil {
		return nil, errors.New("elasticsearch manager is not initialized")
	}
	return defaultManager.GetConn(dbName)
}

// Close 关闭所有 ES 连接
func Close() error {
	if defaultManager == nil {
		return errors.New("elasticsearch manager is not initialized")
	}
	return defaultManager.Close()
}

// HealthCheck ES健康检查
func HealthCheck() map[string]*HealthStatus {
	if defaultManager == nil {
		status := &HealthStatus{
			Status:    "error",
			Error:     "elasticsearch manager is not initialized",
			Timestamp: time.Now().UnixMilli(),
		}
		return map[string]*HealthStatus{"manager": status}
	}
	return defaultManager.HealthCheck()
}
