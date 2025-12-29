package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/logger"
)

// ConfigurableTask 是一个包装器，它使任何 Task 变得可配置。
// 它嵌入了原始的 Task 接口，并覆盖了与调度相关的方法，使其从配置中读取信息。
type ConfigurableTask struct {
	Task              // 嵌入基础任务接口
	config TaskConfig // 持有从配置文件解析的调度信息
}

// NewConfigurableTask 创建一个可配置的任务实例
func NewConfigurableTask(task Task, cfg TaskConfig) *ConfigurableTask {
	return &ConfigurableTask{
		Task:   task,
		config: cfg,
	}
}

// Key 覆盖原始任务的 Key 方法，返回配置中的任务名称
func (ct *ConfigurableTask) Key() string {
	return ct.config.Key
}

// Desc 覆盖原始任务的 Desc 方法，返回配置中的任务描述
func (ct *ConfigurableTask) Desc() string {
	return ct.config.Desc
}

// Spec 覆盖原始任务的 Spec 方法，返回配置中的 CRON 表达式
func (ct *ConfigurableTask) Spec() string {
	return ct.config.Spec
}

// Enabled 覆盖原始任务的 Enabled 方法，返回配置中的启用状态
func (ct *ConfigurableTask) Enabled() bool {
	return ct.config.Enabled
}

// Timeout 覆盖原始任务的 Timeout 方法，返回配置中的超时时间
func (ct *ConfigurableTask) Timeout() time.Duration {
	if ct.config.Timeout == "" {
		return 0 // 默认不超时
	}
	d, err := time.ParseDuration(ct.config.Timeout)
	if err != nil {
		logger.Error(context.Background(), "CRON", fmt.Errorf("invalid timeout duration '%s' for task %s: %v. defaulting to 0", ct.config.Timeout, ct.Key(), err))
		return 0
	}
	return d
}
