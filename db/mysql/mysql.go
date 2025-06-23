package mysql

import (
	"context"
	"fmt"
	"time"

	gocommonlog "github.com/jessewkun/gocommon/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// Init 初始化数据库
func Init() error {
	var initErr error
	for dbName, conf := range Cfgs {
		err := setDefaultConfig(conf)
		if err != nil {
			initErr = fmt.Errorf("mysql %s setDefaultConfig error: %w", dbName, err)
			gocommonlog.ErrorWithMsg(context.Background(), TAGNAME, initErr.Error())
			break
		}
		if err := newClient(dbName, conf); err != nil {
			initErr = fmt.Errorf("connect to mysql %s faild, error: %w", dbName, err)
			gocommonlog.ErrorWithMsg(context.Background(), TAGNAME, initErr.Error())
			break
		}
	}
	return initErr
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	if len(conf.Dsn) < 1 {
		return fmt.Errorf("mysql dsn is invalid")
	}

	if conf.MaxConn == 0 {
		conf.MaxConn = 50
	}

	if conf.MaxIdleConn == 0 {
		conf.MaxIdleConn = 25
	}

	if conf.ConnMaxLife == 0 {
		conf.ConnMaxLife = 3600
	}

	if conf.SlowThreshold == 0 {
		conf.SlowThreshold = 500
	}

	return nil
}

// newClient 连接数据库
func newClient(dbName string, conf *Config) error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	if _, ok := connList.conns[dbName]; ok {
		if connList.conns[dbName] != nil {
			return nil
		}
	}

	logLevel := logger.Silent
	if conf.IsLog {
		logLevel = logger.Info
	}
	slowThreshold := 500 * time.Millisecond
	if conf.SlowThreshold > 0 {
		slowThreshold = time.Duration(conf.SlowThreshold) * time.Millisecond
	}

	var dbOne *gorm.DB
	var err error
	master := conf.Dsn[0]
	slave := conf.Dsn[1:]
	dbOne, err = gorm.Open(mysql.Open(master), &gorm.Config{
		Logger: newMysqlLogger(slowThreshold, logLevel, conf.IgnoreRecordNotFoundError),
	})
	if err != nil {
		return err
	}

	// 配置读写分离
	dbResolverCfg := dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(master)},
		Replicas: []gorm.Dialector{},
		Policy:   dbresolver.RandomPolicy{},
	}

	// 设置从库
	if len(slave) > 0 {
		var replicas []gorm.Dialector
		for i := 0; i < len(slave); i++ {
			replicas = append(replicas, mysql.Open(slave[i]))
		}
		dbResolverCfg.Replicas = replicas
	}

	dbOne.Use(
		dbresolver.Register(dbResolverCfg).
			// SetConnMaxIdleTime(time.Hour).
			SetConnMaxLifetime(time.Duration(conf.ConnMaxLife) * time.Second).
			SetMaxIdleConns(conf.MaxIdleConn).
			SetMaxOpenConns(conf.MaxConn),
	)

	connList.conns[dbName] = dbOne
	gocommonlog.Info(context.Background(), TAGNAME, "connect to mysql %s succ", dbName)

	return nil
}

// GetConn 获取数据库连接
func GetConn(dbIns string) (*gorm.DB, error) {
	connList.mu.RLock()
	defer connList.mu.RUnlock()

	if _, ok := connList.conns[dbIns]; !ok {
		return nil, fmt.Errorf("mysql conn is not found")
	}

	return connList.conns[dbIns], nil
}

// CloseMysql 关闭 MySQL 连接
func Close() error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	var lastErr error
	for dbName, db := range connList.conns {
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				lastErr = fmt.Errorf("get sql.DB for mysql %s failed: %w", dbName, err)
				gocommonlog.ErrorWithMsg(context.Background(), TAGNAME, lastErr.Error())
				continue
			}

			if err := sqlDB.Close(); err != nil {
				lastErr = fmt.Errorf("close mysql %s failed: %w", dbName, err)
				gocommonlog.ErrorWithMsg(context.Background(), TAGNAME, lastErr.Error())
			} else {
				gocommonlog.Info(context.Background(), TAGNAME, "close mysql %s succ", dbName)
			}
		}
	}

	// 清空连接列表
	connList.conns = make(map[string]*gorm.DB)

	return lastErr
}
