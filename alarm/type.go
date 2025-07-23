package alarm

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/config"
	"github.com/spf13/viper"
)

type Config struct {
	Bark    *Bark   `mapstructure:"bark" json:"bark"`
	Feishu  *Feishu `mapstructure:"feishu" json:"feishu"`
	Timeout int     `mapstructure:"timeout" json:"timeout"` // 请求超时时间（秒）
}

func (c *Config) String() string {
	return fmt.Sprintf("bark: %+v, feishu: %+v", c.Bark, c.Feishu)
}

// Reload 重新加载 alarm 配置.
// alarm模块的所有配置项都被认为是安全的，可以进行热更新.
func (c *Config) Reload(v *viper.Viper) {
	if err := v.UnmarshalKey("alarm", c); err != nil {
		fmt.Printf("failed to reload alarm config: %v\n", err)
		return
	}
	fmt.Printf("alarm config reload success, config: %+v\n", c)
}

const TAG = "ALARM"
const MaxRetry = 2

var (
	Cfg    = DefaultConfig()
	client *http.Client
	mu     sync.RWMutex
)

func init() {
	config.Register("alarm", Cfg)
	config.RegisterCallback("alarm", Init)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Bark:    &Bark{BarkIds: []string{}},
		Feishu:  &Feishu{WebhookURL: "", Secret: ""},
		Timeout: 5,
	}
}

// Init 初始化报警系统
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	// 创建 HTTP 客户端
	client = &http.Client{
		Timeout: time.Duration(Cfg.Timeout) * time.Second,
	}

	return nil
}

// getHTTPClient 获取 HTTP 客户端
func getHTTPClient() *http.Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}
