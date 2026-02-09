// Package mongodb 提供 MongoDB 连接管理功能
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const TAG = "MONGODB"

var defaultManager *Manager

// Init 初始化 defaultManager
func Init() error {
	var err error
	defaultManager, err = NewManager(Cfgs)
	return err
}

// setMongoDefaultConfig 设置 MongoDB 默认配置
func setMongoDefaultConfig(conf *Config) error {
	if len(conf.Uris) == 0 {
		return fmt.Errorf("mongodb uris is empty")
	}
	if conf.MaxPoolSize <= 0 {
		conf.MaxPoolSize = 100
	}
	if conf.MinPoolSize <= 0 {
		conf.MinPoolSize = 5
	}
	if conf.MaxConnIdleTime <= 0 {
		conf.MaxConnIdleTime = 300
	}
	if conf.ConnectTimeout <= 0 {
		conf.ConnectTimeout = 10
	}
	if conf.ServerSelectionTimeout <= 0 {
		conf.ServerSelectionTimeout = 5
	}
	if conf.SocketTimeout <= 0 {
		conf.SocketTimeout = 30
	}
	if conf.SlowThreshold <= 0 {
		conf.SlowThreshold = 500
	}
	return nil
}

func buildMongoURI(uris []string) (string, error) {
	if len(uris) == 0 {
		return "", fmt.Errorf("mongodb uris is empty")
	}
	if len(uris) == 1 {
		return uris[0], nil
	}

	// If the first URI contains scheme, try to merge hosts while preserving path/query.
	first := uris[0]
	if strings.HasPrefix(first, "mongodb://") || strings.HasPrefix(first, "mongodb+srv://") {
		schemeSep := strings.Index(first, "://")
		if schemeSep < 0 {
			return first, nil
		}
		scheme := first[:schemeSep+3]
		rest := first[schemeSep+3:]
		pathIdx := strings.Index(rest, "/")
		hostPart := rest
		pathPart := ""
		if pathIdx >= 0 {
			hostPart = rest[:pathIdx]
			pathPart = rest[pathIdx:]
		}

		hosts := []string{hostPart}
		for _, u := range uris[1:] {
			u = strings.TrimPrefix(u, "mongodb://")
			u = strings.TrimPrefix(u, "mongodb+srv://")
			if idx := strings.Index(u, "/"); idx >= 0 {
				u = u[:idx]
			}
			if u != "" {
				hosts = append(hosts, u)
			}
		}

		return scheme + strings.Join(hosts, ",") + pathPart, nil
	}

	// Treat uris as host list without scheme.
	return "mongodb://" + strings.Join(uris, ","), nil
}

// newClient 连接 MongoDB
func newClient(conf *Config) (*mongo.Client, error) {
	uri, err := buildMongoURI(conf.Uris)
	if err != nil {
		return nil, err
	}
	clientOptions := options.Client().ApplyURI(uri)

	// 连接池配置
	clientOptions.SetMaxPoolSize(uint64(conf.MaxPoolSize))
	clientOptions.SetMinPoolSize(uint64(conf.MinPoolSize))
	clientOptions.SetMaxConnIdleTime(time.Duration(conf.MaxConnIdleTime) * time.Second)

	// 超时配置
	clientOptions.SetConnectTimeout(time.Duration(conf.ConnectTimeout) * time.Second)
	clientOptions.SetServerSelectionTimeout(time.Duration(conf.ServerSelectionTimeout) * time.Second)
	clientOptions.SetSocketTimeout(time.Duration(conf.SocketTimeout) * time.Second)

	// 读写配置
	setReadPreference(clientOptions, conf.ReadPreference)
	setWriteConcern(clientOptions, conf.WriteConcern)

	// 注册监控钩子
	if conf.IsLog {
		monitor := newCommandMonitor(time.Duration(conf.SlowThreshold) * time.Millisecond)
		clientOptions.SetMonitor(&event.CommandMonitor{
			Started:   monitor.Started,
			Succeeded: monitor.Succeeded,
			Failed:    monitor.Failed,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ConnectTimeout)*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Duration(conf.ServerSelectionTimeout)*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return client, nil
}

// ... (helper functions for read/write concern)
func setReadPreference(opts *options.ClientOptions, rp string) {
	var pref *readpref.ReadPref
	var err error
	switch rp {
	case "primary":
		pref = readpref.Primary()
	case "primaryPreferred":
		pref = readpref.PrimaryPreferred()
	case "secondary":
		pref = readpref.Secondary()
	case "secondaryPreferred":
		pref = readpref.SecondaryPreferred()
	case "nearest":
		pref = readpref.Nearest()
	default:
		pref = readpref.Primary()
	}
	if err == nil {
		opts.SetReadPreference(pref)
	}
}

func setWriteConcern(opts *options.ClientOptions, wc string) {
	var concern *writeconcern.WriteConcern
	switch wc {
	case "majority":
		concern = writeconcern.Majority()
	case "1":
		concern = writeconcern.W1()
	case "0":
		concern = writeconcern.Unacknowledged()
	default:
		concern = writeconcern.Majority()
	}
	opts.SetWriteConcern(concern)
}

// GetConn 获取 MongoDB 客户端
func GetConn(dbIns string) (*mongo.Client, error) {
	if defaultManager == nil {
		return nil, errors.New("mongodb manager is not initialized")
	}
	return defaultManager.GetConn(dbIns)
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
	if defaultManager == nil {
		return errors.New("mongodb manager is not initialized")
	}
	return defaultManager.Close()
}

// HealthCheck MongoDB 健康检查
func HealthCheck() map[string]*HealthStatus {
	if defaultManager == nil {
		status := &HealthStatus{
			Status:    "error",
			Error:     "mongodb manager is not initialized",
			Timestamp: time.Now().UnixMilli(),
		}
		return map[string]*HealthStatus{"manager": status}
	}
	return defaultManager.HealthCheck()
}

// WithTransaction 使用事务执行操作
func WithTransaction(client *mongo.Client, fn func(mongo.SessionContext) error) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	// 事务超时建议从配置中读取或作为参数传入
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err = sessCtx.StartTransaction(options.Transaction().SetReadConcern(readconcern.Snapshot())); err != nil {
			return err
		}
		if err = fn(sessCtx); err != nil {
			// If the operation fails, abort the transaction but return the original error.
			// The abort error is logged if it occurs but not returned to the caller.
			if abortErr := sessCtx.AbortTransaction(sessCtx); abortErr != nil {
				logger.ErrorWithMsg(sessCtx, TAG, "failed to abort transaction: %v", abortErr)
			}
			return err
		}
		return sessCtx.CommitTransaction(sessCtx)
	})
}
