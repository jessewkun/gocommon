// Package oss 提供阿里云OSS的客户端
package oss

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OssConfig OSS配置选项
type OssConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	// 连接池配置
	MaxConnections    int
	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration
	// 重试配置
	MaxRetries int
	RetryDelay time.Duration
}

// Oss OSS客户端结构体
type Oss struct {
	config     *OssConfig
	client     *oss.Client
	clientOnce sync.Once
	clientErr  error
	mu         sync.RWMutex
}

// NewOss 创建新的OSS客户端
func NewOss(config *OssConfig) (*Oss, error) {
	if config == nil {
		return nil, errors.New("配置不能为空")
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 设置默认值
	if config.MaxConnections <= 0 {
		config.MaxConnections = 100
	}
	if config.ConnectionTimeout <= 0 {
		config.ConnectionTimeout = 30 * time.Second
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 60 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &Oss{
		config: config,
	}, nil
}

// NewOssSimple 创建OSS客户端的简化方法
func NewOssSimple(endpoint, accessKeyID, accessKeySecret string) (*Oss, error) {
	config := &OssConfig{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	}
	return NewOss(config)
}

// validateConfig 验证配置参数
func validateConfig(config *OssConfig) error {
	if config.Endpoint == "" {
		return errors.New("Endpoint不能为空")
	}
	if config.AccessKeyID == "" {
		return errors.New("AccessKeyID不能为空")
	}
	if config.AccessKeySecret == "" {
		return errors.New("AccessKeySecret不能为空")
	}
	return nil
}

// getClient 获取OSS客户端（单例模式）
func (o *Oss) newClient() error {
	o.clientOnce.Do(func() {
		client, err := oss.New(
			o.config.Endpoint,
			o.config.AccessKeyID,
			o.config.AccessKeySecret,
		)
		if err != nil {
			o.clientErr = err
			return
		}
		o.client = client
	})

	if o.clientErr != nil {
		return o.clientErr
	}

	return nil
}

// Close 关闭OSS客户端连接
func (o *Oss) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.client != nil {
		// OSS客户端没有显式的Close方法，这里主要是清理引用
		o.client = nil
	}
	return nil
}
