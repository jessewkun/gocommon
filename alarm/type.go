package alarm

import (
	"context"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/config"
	"github.com/jessewkun/gocommon/http"
	"github.com/jessewkun/gocommon/logger"
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
// alarm 模块的所有配置项都被认为是安全的，可以进行热更新.
func (c *Config) Reload(v *viper.Viper) error {
	if err := v.UnmarshalKey("alarm", c); err != nil {
		logger.ErrorWithMsg(context.Background(), "ALARM", "failed to reload alarm config: %v", err)
		return err
	}
	logger.Info(context.Background(), "ALARM", "alarm config reload success, config: %+v", c)
	// 重新初始化http客户端以应用新的超时设置
	Init()
	return nil
}

const TAG = "ALARM"
const MaxRetry = 2

var (
	Cfg             = DefaultConfig()
	alarmHTTPClient *http.Client // 使用 @http 模块的客户端
)

func init() {
	config.Register("alarm", Cfg)
	config.RegisterCallback("alarm", Init, "config", "http", "log")
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
	isLog := false
	alarmHTTPClient = http.NewClient(http.Option{
		Timeout: time.Duration(Cfg.Timeout) * time.Second,
		Retry:   MaxRetry,
		IsLog:   &isLog, // 报警模块不记录自己的请求日志，避免循环
	})

	return nil
}
