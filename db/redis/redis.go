// Package redis 提供 Redis 连接管理功能
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

const TAG = "REDIS"

// Init 初始化 defaultManager
func Init() error {
	var err error
	defaultManager, err = NewManager(Cfgs)
	return err
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	if len(conf.Addrs) < 1 {
		return errors.New("redis addrs is empty")
	}
	if conf.PoolSize <= 0 {
		conf.PoolSize = 100
	}
	if conf.IdleTimeout <= 0 {
		conf.IdleTimeout = 300 // 5 minutes
	}
	if conf.IdleCheckFrequency <= 0 {
		conf.IdleCheckFrequency = 10
	}
	if conf.MinIdleConns <= 0 {
		conf.MinIdleConns = 3
	}
	if conf.MaxRetries <= 0 {
		conf.MaxRetries = 3
	}
	if conf.DialTimeout <= 0 {
		conf.DialTimeout = 2
	}
	if conf.SlowThreshold <= 0 {
		conf.SlowThreshold = 200
	}

	return nil
}

// newClient 根据配置连接 redis，返回一个通用客户端
func newClient(conf *Config) (redis.UniversalClient, error) {
	var client redis.UniversalClient

	if conf.IsCluster {
		// 集群模式
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:              conf.Addrs,
			Password:           conf.Password,
			PoolSize:           conf.PoolSize,
			IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
			MinIdleConns:       conf.MinIdleConns,
			MaxRetries:         conf.MaxRetries,
			DialTimeout:        time.Duration(conf.DialTimeout) * time.Second,
		})
	} else {
		// 单点模式
		client = redis.NewClient(&redis.Options{
			Addr:               conf.Addrs[0], // 单点模式只取第一个地址
			Password:           conf.Password,
			DB:                 conf.Db,
			PoolSize:           conf.PoolSize,
			IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
			MinIdleConns:       conf.MinIdleConns,
			MaxRetries:         conf.MaxRetries,
			DialTimeout:        time.Duration(conf.DialTimeout) * time.Second,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DialTimeout)*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()

	if conf.IsLog {
		client.AddHook(newRedisHook(time.Duration(conf.SlowThreshold) * time.Millisecond))
	}

	return client, err
}

// GetConn 获得redis连接
func GetConn(dbIns string) (redis.UniversalClient, error) {
	if defaultManager == nil {
		return nil, errors.New("redis manager is not initialized")
	}
	return defaultManager.GetConn(dbIns)
}

// Close 关闭 Redis 连接
func Close() error {
	if defaultManager == nil {
		return errors.New("redis manager is not initialized")
	}
	return defaultManager.Close()
}

// HealthCheck redis健康检查
func HealthCheck() map[string]*HealthStatus {
	if defaultManager == nil {
		status := &HealthStatus{
			Status:    "error",
			Error:     "redis manager is not initialized",
			Timestamp: time.Now().UnixMilli(),
		}
		return map[string]*HealthStatus{"manager": status}
	}
	return defaultManager.HealthCheck()
}
