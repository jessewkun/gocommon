// Package nacos 提供 Nacos 连接管理功能
package nacos

import (
	"errors"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Init 初始化 defaultManager
func Init() error {
	var err error
	defaultManager, err = NewManager(Cfgs)
	return err
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	defaultConf := DefaultConfig()
	if conf.Host == "" {
		conf.Host = defaultConf.Host
	}
	if conf.Port == 0 {
		conf.Port = defaultConf.Port
	}
	if conf.Namespace == "" {
		conf.Namespace = defaultConf.Namespace
	}
	if conf.Group == "" {
		conf.Group = defaultConf.Group
	}
	if conf.Timeout == 0 {
		conf.Timeout = defaultConf.Timeout
	}
	if conf.LogDir == "" {
		conf.LogDir = defaultConf.LogDir
	}
	if conf.CacheDir == "" {
		conf.CacheDir = defaultConf.CacheDir
	}
	if conf.LogLevel == "" {
		conf.LogLevel = defaultConf.LogLevel
	}
	return nil
}

// newClient 根据配置创建 nacos 客户端
func newClient(conf *Config) (*Client, error) {
	// 创建服务器配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: conf.Host,
			Port:   conf.Port,
		},
	}

	// 创建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         conf.Namespace,
		TimeoutMs:           uint64(conf.Timeout),
		NotLoadCacheAtStart: conf.NotLoadCacheAtStart,
		LogDir:              conf.LogDir,
		CacheDir:            conf.CacheDir,
		LogLevel:            conf.LogLevel,
	}

	// 如果有用户名密码，设置认证信息
	if conf.Username != "" {
		clientConfig.Username = conf.Username
		clientConfig.Password = conf.Password
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	// 创建命名客户端
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		configClient.CloseClient() // 如果命名客户端创建失败，关闭已创建的配置客户端
		return nil, fmt.Errorf("failed to create naming client: %w", err)
	}

	client := &Client{
		configClient: configClient,
		namingClient: namingClient,
		config:       conf,
	}

	return client, nil
}

// GetClient 获取nacos客户端连接
func GetClient(clientName string) (*Client, error) {
	if defaultManager == nil {
		return nil, errors.New("nacos manager is not initialized")
	}
	return defaultManager.GetClient(clientName)
}

// Close 关闭所有nacos客户端连接
func Close() error {
	if defaultManager == nil {
		return errors.New("nacos manager is not initialized")
	}
	return defaultManager.Close()
}
