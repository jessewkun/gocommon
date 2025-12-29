package cron

// TaskConfig 保存单个定时任务的通用配置结构
type TaskConfig struct {
	Key     string `mapstructure:"key"`     // 任务标识
	Desc    string `mapstructure:"desc"`    // 任务描述
	Spec    string `mapstructure:"spec"`    // CRON表达式
	Enabled bool   `mapstructure:"enabled"` // 是否启用
	Timeout string `mapstructure:"timeout"` // 超时时间，例如 "5m", "1h"
}
