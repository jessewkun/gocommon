package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const TAG = "MONGODB"

type Connections struct {
	mu    sync.RWMutex
	conns map[string]*mongo.Client
}

var connList = &Connections{
	conns: make(map[string]*mongo.Client),
}

// Init 初始化 MongoDB 连接
func Init() error {
	var initErr error
	for dbName, conf := range Cfgs {
		err := setMongoDefaultConfig(conf)
		if err != nil {
			initErr = fmt.Errorf("mongodb %s setDefaultConfig error: %w", dbName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
		if err := newClient(dbName, conf); err != nil {
			initErr = fmt.Errorf("connect to mongodb %s failed, error: %w", dbName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
	}
	return initErr
}

// setMongoDefaultConfig 设置 MongoDB 默认配置
func setMongoDefaultConfig(conf *Config) error {
	if len(conf.Uris) < 1 {
		return fmt.Errorf("mongodb uris is invalid")
	}

	if conf.MaxPoolSize == 0 {
		conf.MaxPoolSize = 100
	}

	if conf.MinPoolSize == 0 {
		conf.MinPoolSize = 5
	}

	if conf.MaxConnIdleTime == 0 {
		conf.MaxConnIdleTime = 300
	}

	if conf.ConnectTimeout == 0 {
		conf.ConnectTimeout = 10
	}

	if conf.ServerSelectionTimeout == 0 {
		conf.ServerSelectionTimeout = 5
	}

	if conf.SocketTimeout == 0 {
		conf.SocketTimeout = 30
	}

	return nil
}

// newClient 连接 MongoDB
func newClient(dbName string, conf *Config) error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	if _, ok := connList.conns[dbName]; ok {
		if connList.conns[dbName] != nil {
			return nil
		}
	}

	// 构建连接选项
	clientOptions := options.Client()

	// 设置连接字符串
	clientOptions.ApplyURI(conf.Uris[0])

	// 设置连接池配置
	clientOptions.SetMaxPoolSize(uint64(conf.MaxPoolSize))
	clientOptions.SetMinPoolSize(uint64(conf.MinPoolSize))
	clientOptions.SetMaxConnIdleTime(time.Duration(conf.MaxConnIdleTime) * time.Second)

	// 设置超时配置
	clientOptions.SetConnectTimeout(time.Duration(conf.ConnectTimeout) * time.Second)
	clientOptions.SetServerSelectionTimeout(time.Duration(conf.ServerSelectionTimeout) * time.Second)
	clientOptions.SetSocketTimeout(time.Duration(conf.SocketTimeout) * time.Second)

	// 设置读取偏好
	if conf.ReadPreference != "" {
		switch conf.ReadPreference {
		case "primary":
			clientOptions.SetReadPreference(readpref.Primary())
		case "primaryPreferred":
			clientOptions.SetReadPreference(readpref.PrimaryPreferred())
		case "secondary":
			clientOptions.SetReadPreference(readpref.Secondary())
		case "secondaryPreferred":
			clientOptions.SetReadPreference(readpref.SecondaryPreferred())
		case "nearest":
			clientOptions.SetReadPreference(readpref.Nearest())
		default:
			clientOptions.SetReadPreference(readpref.Primary())
		}
	} else {
		clientOptions.SetReadPreference(readpref.Primary())
	}

	// 设置写入关注
	if conf.WriteConcern != "" {
		switch conf.WriteConcern {
		case "majority":
			clientOptions.SetWriteConcern(writeconcern.Majority())
		case "1":
			clientOptions.SetWriteConcern(writeconcern.W1())
		case "0":
			clientOptions.SetWriteConcern(writeconcern.Unacknowledged())
		default:
			clientOptions.SetWriteConcern(writeconcern.Majority())
		}
	} else {
		clientOptions.SetWriteConcern(writeconcern.Majority())
	}

	// 创建客户端
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongodb: %w", err)
	}

	connList.conns[dbName] = client
	logger.Info(context.Background(), TAG, "connect to mongodb %s succ", dbName)

	return nil
}

// GetConn 获取 MongoDB 客户端
func GetConn(dbIns string) (*mongo.Client, error) {
	connList.mu.RLock()
	defer connList.mu.RUnlock()

	if _, ok := connList.conns[dbIns]; !ok {
		return nil, fmt.Errorf("mongodb client is not found")
	}

	return connList.conns[dbIns], nil
}

// GetDatabase 获取 MongoDB 数据库实例
func GetDatabase(dbIns, databaseName string) (*mongo.Database, error) {
	client, err := GetConn(dbIns)
	if err != nil {
		return nil, err
	}

	return client.Database(databaseName), nil
}

// GetCollection 获取 MongoDB 集合实例
func GetCollection(dbIns, databaseName, collectionName string) (*mongo.Collection, error) {
	database, err := GetDatabase(dbIns, databaseName)
	if err != nil {
		return nil, err
	}

	return database.Collection(collectionName), nil
}

// Close 关闭 MongoDB 连接
func Close() error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	var lastErr error
	for dbName, client := range connList.conns {
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := client.Disconnect(ctx); err != nil {
				lastErr = fmt.Errorf("disconnect mongodb %s failed: %w", dbName, err)
				logger.ErrorWithMsg(context.Background(), TAG, lastErr.Error())
			} else {
				logger.Info(context.Background(), TAG, "disconnect mongodb %s succ", dbName)
			}
			cancel()
		}
	}

	// 清空连接列表
	connList.conns = make(map[string]*mongo.Client)

	return lastErr
}
