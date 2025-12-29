package nacos

import (
	"github.com/jessewkun/gocommon/config"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
)

// Client nacos客户端
type Client struct {
	configClient config_client.IConfigClient
	namingClient naming_client.INamingClient
	config       *Config
}

// Close 关闭客户端，释放资源
func (c *Client) Close() error {
	if c.configClient != nil {
		c.configClient.CloseClient()
	}
	if c.namingClient != nil {
		c.namingClient.CloseClient()
	}
	return nil
}

// Config nacos配置
type Config struct {
	Host                string `mapstructure:"host" json:"host"`                                       // nacos服务器地址
	Port                uint64 `mapstructure:"port" json:"port"`                                       // nacos服务器端口
	Namespace           string `mapstructure:"namespace" json:"namespace"`                             // 命名空间
	Group               string `mapstructure:"group" json:"group"`                                     // 配置组
	Username            string `mapstructure:"username" json:"username"`                               // 用户名
	Password            string `mapstructure:"password" json:"password"`                               // 密码
	Timeout             int    `mapstructure:"timeout" json:"timeout"`                                 // 超时时间(毫秒)
	NotLoadCacheAtStart bool   `mapstructure:"not_load_cache_at_start" json:"not_load_cache_at_start"` // 是否在启动时加载缓存
	LogDir              string `mapstructure:"log_dir" json:"log_dir"`                                 // 日志存储路径
	CacheDir            string `mapstructure:"cache_dir" json:"cache_dir"`                             // 缓存存储路径
	LogLevel            string `mapstructure:"log_level" json:"log_level"`                             // 日志级别
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:                "localhost",
		Port:                8848,
		Namespace:           "public",
		Group:               "DEFAULT_GROUP",
		Username:            "",
		Password:            "",
		Timeout:             5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "warn",
	}
}

// Cfgs 全局配置管理
var Cfgs = make(map[string]*Config)

var defaultManager *Manager

const TAG = "NACOS"

func init() {
	config.Register("nacos", &Cfgs)
	config.RegisterCallback("nacos", Init, "config", "log")
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ServiceName string            `json:"service_name"`
	IP          string            `json:"ip"`
	Port        uint64            `json:"port"`
	Weight      float64           `json:"weight"`
	Enable      bool              `json:"enable"`
	Healthy     bool              `json:"healthy"`
	Metadata    map[string]string `json:"metadata"`
}

// ConfigInfo 配置信息
type ConfigInfo struct {
	DataID  string `json:"data_id"`
	Group   string `json:"group"`
	Content string `json:"content"`
}

// ServiceChangeEvent 服务变化事件
type ServiceChangeEvent struct {
	ServiceName string        `json:"service_name"`
	Instances   []ServiceInfo `json:"instances"`
}

// ConfigChangeEvent 配置变化事件
type ConfigChangeEvent struct {
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	DataID    string `json:"data_id"`
	Content   string `json:"content"`
}
