package elasticsearch

import (
	"context"
	"fmt"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	gocommonlog "github.com/jessewkun/gocommon/logger"
)

const TAG = "ELASTICSEARCH"

type Connections struct {
	mu    sync.RWMutex
	conns map[string]*Client
}

var connList = &Connections{
	conns: make(map[string]*Client),
}

func Init() error {
	var initErr error
	for dbName, conf := range Cfgs {
		if err := newClient(dbName, conf); err != nil {
			initErr = fmt.Errorf("connect to elasticsearch %s faild, error: %w", dbName, err)
			gocommonlog.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
	}
	return initErr
}

// GetConn 获取指定名称的 ES 客户端
func GetConn(dbName string) (*Client, error) {
	connList.mu.RLock()
	defer connList.mu.RUnlock()

	if client, ok := connList.conns[dbName]; ok && client != nil {
		return client, nil
	}

	return nil, fmt.Errorf("elasticsearch client '%s' not found", dbName)
}

// newClient 创建新的 ES 客户端
func newClient(dbName string, cfg *Config) error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	if _, ok := connList.conns[dbName]; ok {
		if connList.conns[dbName] != nil {
			return nil
		}
	}

	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return err
	}

	connList.conns[dbName] = &Client{ES: es}
	gocommonlog.Info(context.Background(), TAG, "connect to elasticsearch %s succ", dbName)

	return nil
}
