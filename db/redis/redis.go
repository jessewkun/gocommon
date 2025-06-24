package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/utils"

	"github.com/go-redis/redis/v8"
)

const TAG = "REDIS"

type Connections struct {
	mu    sync.RWMutex
	conns map[string]map[string]*redis.Client
}

var connList = &Connections{
	conns: make(map[string]map[string]*redis.Client),
}

// Init 初始化redis，使用模块内注册的Cfgs
func Init() error {
	var initErr error
	for dbName, conf := range Cfgs {
		if err := setDefaultConfig(conf); err != nil {
			initErr = fmt.Errorf("redis %s setDefaultConfig error: %w", dbName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
		if err := newClient(dbName, conf); err != nil {
			initErr = fmt.Errorf("connect to redis %s error: %w", dbName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
	}
	return initErr
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	if len(conf.Addrs) < 1 {
		return errors.New("redis addrs is empty")
	}
	if conf.PoolSize == 0 {
		conf.PoolSize = 500
	}
	if conf.IdleTimeout == 0 {
		conf.IdleTimeout = 1
	}
	if conf.IdleCheckFrequency == 0 {
		conf.IdleCheckFrequency = 10
	}
	if conf.MinIdleConns == 0 {
		conf.MinIdleConns = 3
	}
	if conf.MaxRetries == 0 {
		conf.MaxRetries = 3
	}
	if conf.DialTimeout == 0 {
		conf.DialTimeout = 2
	}
	if conf.SlowThreshold == 0 {
		conf.SlowThreshold = 200
	}

	return nil
}

// newClient 连接 redis
func newClient(dbName string, conf *Config) error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	if _, ok := connList.conns[dbName]; ok {
		if connList.conns[dbName] != nil {
			return nil
		}
	}

	connList.conns[dbName] = make(map[string]*redis.Client, 0)
	for _, addr := range conf.Addrs {
		client := redis.NewClient(&redis.Options{
			Addr:               addr,
			Password:           conf.Password,
			DB:                 conf.Db,
			PoolSize:           conf.PoolSize,
			IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
			IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
			MinIdleConns:       conf.MinIdleConns,
			MaxRetries:         conf.MaxRetries,
			DialTimeout:        time.Duration(conf.DialTimeout) * time.Second,
		})
		if conf.IsLog {
			client.AddHook(newRedisHook(time.Duration(conf.SlowThreshold) * time.Millisecond))
		}
		connList.conns[dbName][addr] = client
		logger.Info(context.Background(), TAG, "connect to redis %s addr %s succ", dbName, addr)
	}
	return nil
}

// GetConn 获得redis连接
func GetConn(dbIns string) (*redis.Client, error) {
	connList.mu.RLock()
	defer connList.mu.RUnlock()

	if len(connList.conns) < 1 {
		return nil, errors.New("redis connList is empty")
	}

	conns, ok := connList.conns[dbIns]
	if !ok {
		return nil, errors.New("redis conn is not found")
	}

	keys := make([]string, 0, len(conns))
	for key := range conns {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil, errors.New("redis conn is empty")
	}

	randomKey := keys[utils.RandomNum(0, len(keys)-1)]
	return conns[randomKey], nil
}

// Close 关闭 Redis 连接
func Close() error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	var lastErr error
	for dbName, conns := range connList.conns {
		for addr, conn := range conns {
			if conn != nil {
				if err := conn.Close(); err != nil {
					lastErr = fmt.Errorf("close redis %s addr %s failed: %w", dbName, addr, err)
					logger.ErrorWithMsg(context.Background(), TAG, lastErr.Error())
				} else {
					logger.Info(context.Background(), TAG, "close redis %s addr %s succ", dbName, addr)
				}
			}
		}
	}

	// 清空连接列表
	connList.conns = make(map[string]map[string]*redis.Client)

	return lastErr
}
