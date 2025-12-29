package cron

import (
	"context"
	"time"
)

// BaseTask 定时任务基类，提供通用方法的默认实现
type BaseTask struct {
}

// Key 获取任务标识
func (t *BaseTask) Key() string {
	return ""
}

// Desc 获取任务描述
func (t *BaseTask) Desc() string {
	return ""
}

// Spec 默认实现，实际值由 ConfigurableTask 提供
func (t *BaseTask) Spec() string {
	return ""
}

// Enabled 默认实现，实际值由 ConfigurableTask 提供
func (t *BaseTask) Enabled() bool {
	return false
}

// Timeout 默认实现，实际值由 ConfigurableTask 提供
func (t *BaseTask) Timeout() time.Duration {
	return 0
}

// BeforeRun 任务执行前，提供默认空实现
func (t *BaseTask) BeforeRun(ctx context.Context) error {
	return nil
}

// Run 执行任务，提供默认空实现
func (t *BaseTask) Run(ctx context.Context) error {
	return nil
}

// AfterRun 任务执行后，提供默认空实现
func (t *BaseTask) AfterRun(ctx context.Context) error {
	return nil
}
