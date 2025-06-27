package nacos

import (
	"context"
	"fmt"

	"github.com/jessewkun/gocommon/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Init 初始化nacos客户端
func Init() error {
	var initErr error
	for clientName, conf := range Cfgs {
		err := setDefaultConfig(conf)
		if err != nil {
			initErr = fmt.Errorf("nacos %s setDefaultConfig error: %w", clientName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
		if err := newClient(clientName, conf); err != nil {
			initErr = fmt.Errorf("connect to nacos %s failed, error: %w", clientName, err)
			logger.ErrorWithMsg(context.Background(), TAG, initErr.Error())
			break
		}
	}
	return initErr
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(conf *Config) error {
	if conf.Host == "" {
		conf.Host = "localhost"
	}

	if conf.Port == 0 {
		conf.Port = 8848
	}

	if conf.Namespace == "" {
		conf.Namespace = "public"
	}

	if conf.Group == "" {
		conf.Group = "DEFAULT_GROUP"
	}

	if conf.Timeout == 0 {
		conf.Timeout = 5000
	}

	return nil
}

// newClient 创建nacos客户端
func newClient(clientName string, conf *Config) error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	if _, ok := connList.clients[clientName]; ok {
		if connList.clients[clientName] != nil {
			return nil
		}
	}

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
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
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
		return fmt.Errorf("failed to create config client: %w", err)
	}

	// 创建命名客户端
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create naming client: %w", err)
	}

	client := &Client{
		configClient: configClient,
		namingClient: namingClient,
		config:       conf,
	}

	connList.clients[clientName] = client
	logger.Info(context.Background(), TAG, "connect to nacos %s succ", clientName)

	return nil
}

// GetConn 获取nacos客户端连接
func GetConn(clientName string) (*Client, error) {
	connList.mu.RLock()
	defer connList.mu.RUnlock()

	if _, ok := connList.clients[clientName]; !ok {
		return nil, fmt.Errorf("nacos client '%s' is not found", clientName)
	}

	return connList.clients[clientName], nil
}

// Close 关闭所有nacos客户端连接
func Close() error {
	connList.mu.Lock()
	defer connList.mu.Unlock()

	var lastErr error
	for clientName, client := range connList.clients {
		if client != nil {
			if err := client.Close(); err != nil {
				lastErr = fmt.Errorf("close nacos %s failed: %w", clientName, err)
				logger.ErrorWithMsg(context.Background(), TAG, lastErr.Error())
			} else {
				logger.Info(context.Background(), TAG, "close nacos %s succ", clientName)
			}
		}
	}

	// 清空连接列表
	connList.clients = make(map[string]*Client)

	return lastErr
}
