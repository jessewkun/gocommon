package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/utils"

	"github.com/go-redis/redis/v8"
)

const TAGNAME = "REDIS"

var connList map[string]map[string]*redis.Client

func init() {
	connList = make(map[string]map[string]*redis.Client)
}

// InitRedis 初始化redis
func InitRedis(cfg map[string]*Config) {
	for dbName, conf := range cfg {
		err := setDefaultConfig(conf)
		if err != nil {
			logger.ErrorWithMsg(context.Background(), TAGNAME, "mysql %s setDefaultConfig error: %s", dbName, err)
			continue
		}
		if err := redisConnect(dbName, conf); err != nil {
			logger.ErrorWithMsg(context.Background(), TAGNAME, "connect to redis %s error %s", dbName, err)
			continue
		}
	}
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	if len(conf.Addrs) < 1 {
		return errors.New(fmt.Sprintf("redis addrs is empty"))
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

	return nil
}

// redisConnect 连接 redis
func redisConnect(dbName string, conf *Config) error {
	connList[dbName] = make(map[string]*redis.Client, 0)
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
			client.AddHook(&RedisHook{})
		}
		connList[dbName][addr] = client
		logger.Info(context.Background(), TAGNAME, "connect to redis %s addr %s succ", dbName, addr)
	}
	return nil
}

// GetConn 获得redis
func GetConn(dbIns string) *redis.Client {
	if len(connList) < 1 {
		return nil
	}
	if _, ok := connList[dbIns]; !ok {
		return nil
	}

	keys := make([]string, 0, len(connList[dbIns]))
	for key := range connList[dbIns] {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}

	randomKey := keys[utils.RandomNum(0, len(keys)-1)]
	return connList[dbIns][randomKey]
}

// redis health check
func HealthCheck() map[string]map[string]string {
	resp := make(map[string]map[string]string)
	for dbName, conns := range connList {
		resp[dbName] = make(map[string]string)
		for addr, conn := range conns {
			if _, err := conn.Ping(context.Background()).Result(); err != nil {
				logger.ErrorWithMsg(context.Background(), TAGNAME, "redis ping db %s addr %s error %s", dbName, addr, err)
				resp[dbName][addr] = err.Error()
			} else {
				resp[dbName][addr] = "succ"
			}
		}
	}
	return resp
}
