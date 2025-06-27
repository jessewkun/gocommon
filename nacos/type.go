package nacos

import (
	"sync"

	"github.com/jessewkun/gocommon/config"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
)

// Client nacos客户端
type Client struct {
	configClient config_client.IConfigClient
	namingClient naming_client.INamingClient
	config       *Config
	mu           sync.RWMutex
}

// Config nacos配置
type Config struct {
	Host      string `mapstructure:"host" json:"host"`           // nacos服务器地址
	Port      uint64 `mapstructure:"port" json:"port"`           // nacos服务器端口
	Namespace string `mapstructure:"namespace" json:"namespace"` // 命名空间
	Group     string `mapstructure:"group" json:"group"`         // 配置组
	Username  string `mapstructure:"username" json:"username"`   // 用户名
	Password  string `mapstructure:"password" json:"password"`   // 密码
	Timeout   int    `mapstructure:"timeout" json:"timeout"`     // 超时时间(毫秒)
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:      "localhost",
		Port:      8848,
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
		Username:  "",
		Password:  "",
		Timeout:   5000,
	}
}

// 全局配置管理
var Cfgs = make(map[string]*Config)

const TAG = "NACOS"

// Connections 连接管理器
type Connections struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

var connList = &Connections{
	clients: make(map[string]*Client),
}

func init() {
	config.Register("nacos", &Cfgs)
	config.RegisterCallback("nacos", Init)
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
	DataId  string `json:"data_id"`
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
	DataId    string `json:"data_id"`
	Content   string `json:"content"`
}
