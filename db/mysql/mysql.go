// Package mysql 提供 MySQL 数据库连接管理功能
package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// Init 初始化 defaultManager
func Init() error {
	var err error
	defaultManager, err = NewManager(Cfgs)
	return err
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
	if conf.ConnMaxLifeTime == 0 {
		conf.ConnMaxLifeTime = 3600
	}
	if conf.ConnMaxIdleTime == 0 {
		conf.ConnMaxIdleTime = 600
	}
	if conf.SlowThreshold == 0 {
		conf.SlowThreshold = 500
	}
	return nil
}

// newClient 连接数据库
func newClient(conf *Config) (*gorm.DB, error) {
	// 解析日志级别
	logLevel := logger.Silent
	switch conf.LogLevel {
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	}

	slowThreshold := 500 * time.Millisecond
	if conf.SlowThreshold > 0 {
		slowThreshold = time.Duration(conf.SlowThreshold) * time.Millisecond
	}

	master := conf.Dsn[0]
	slave := conf.Dsn[1:]
	dbOne, err := gorm.Open(mysql.Open(master), &gorm.Config{
		Logger: newMysqlLogger(slowThreshold, logLevel, conf.IgnoreRecordNotFoundError),
	})
	if err != nil {
		return nil, err
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

	err = dbOne.Use(
		dbresolver.Register(dbResolverCfg).
			SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second).
			SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second).
			SetMaxIdleConns(conf.MaxIdleConn).
			SetMaxOpenConns(conf.MaxConn),
	)
	if err != nil {
		return nil, err
	}

	return dbOne, nil
}

// GetConn 获取数据库连接
func GetConn(dbIns string) (*gorm.DB, error) {
	if defaultManager == nil {
		return nil, fmt.Errorf("mysql manager is not initialized")
	}
	return defaultManager.GetConn(dbIns)
}

// Close 关闭数据库连接
func Close() error {
	if defaultManager == nil {
		return fmt.Errorf("mysql manager is not initialized")
	}
	return defaultManager.Close()
}

// HealthCheck 健康检查
func HealthCheck() map[string]*HealthStatus {
	if defaultManager == nil {
		status := &HealthStatus{
			Status:    "error",
			Error:     "mysql manager is not initialized",
			Timestamp: time.Now().UnixMilli(),
		}
		return map[string]*HealthStatus{"manager": status}
	}
	return defaultManager.HealthCheck()
}
