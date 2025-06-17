package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

// MongoTransaction MongoDB 事务
type MongoTransaction struct {
	client    *mongo.Client
	session   mongo.Session
	ctx       context.Context
	cancelCtx context.CancelFunc
}

// NewMongoTransaction 创建新的 MongoDB 事务
func NewMongoTransaction(client *mongo.Client) (*MongoTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	session, err := client.StartSession()
	if err != nil {
		cancel()
		return nil, err
	}

	return &MongoTransaction{
		client:    client,
		session:   session,
		ctx:       ctx,
		cancelCtx: cancel,
	}, nil
}

// StartTransaction 开始事务
func (t *MongoTransaction) StartTransaction() error {
	return mongo.WithSession(t.ctx, t.session, func(sessCtx mongo.SessionContext) error {
		return sessCtx.StartTransaction(options.Transaction().SetReadConcern(readconcern.Snapshot()))
	})
}

// Commit 提交事务
func (t *MongoTransaction) Commit() error {
	defer t.cleanup()
	return mongo.WithSession(t.ctx, t.session, func(sessCtx mongo.SessionContext) error {
		return sessCtx.CommitTransaction(sessCtx)
	})
}

// Rollback 回滚事务
func (t *MongoTransaction) Rollback() error {
	defer t.cleanup()
	return mongo.WithSession(t.ctx, t.session, func(sessCtx mongo.SessionContext) error {
		return sessCtx.AbortTransaction(sessCtx)
	})
}

// GetSession 获取会话上下文
func (t *MongoTransaction) GetSession() mongo.Session {
	return t.session
}

// GetContext 获取事务上下文
func (t *MongoTransaction) GetContext() context.Context {
	return t.ctx
}

// cleanup 清理资源
func (t *MongoTransaction) cleanup() {
	if t.session != nil {
		t.session.EndSession(t.ctx)
	}
	if t.cancelCtx != nil {
		t.cancelCtx()
	}
}

// WithTransaction 使用事务执行操作
func WithTransaction(client *mongo.Client, fn func(mongo.SessionContext) error) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := sessCtx.StartTransaction(options.Transaction().SetReadConcern(readconcern.Snapshot())); err != nil {
			return err
		}

		if err := fn(sessCtx); err != nil {
			if abortErr := sessCtx.AbortTransaction(sessCtx); abortErr != nil {
				return abortErr
			}
			return err
		}

		return sessCtx.CommitTransaction(sessCtx)
	})
}
