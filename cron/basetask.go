package cron

import (
	"context"
	"time"
)

// BaseTask 定时任务基类
type BaseTask struct {
	TaskName    string        // 任务名称
	TaskDesc    string        // 任务描述
	TaskEnabled bool          // 任务是否启用
	TaskSpec    string        // 任务调度表达式
	TaskTimeout time.Duration // 任务超时时间
}

// Name 获取任务名称
func (t *BaseTask) Name() string {
	return t.TaskName
}

// Desc 获取任务描述
func (t *BaseTask) Desc() string {
	return t.TaskDesc
}

// Spec 获取任务调度表达式
func (t *BaseTask) Spec() string {
	return t.TaskSpec
}

// Enabled 获取是否启用
func (t *BaseTask) Enabled() bool {
	return t.TaskEnabled
}

// Timeout 获取任务超时时间
func (t *BaseTask) Timeout() time.Duration {
	return t.TaskTimeout
}

// BeforeRun 任务执行前
func (t *BaseTask) BeforeRun(ctx context.Context) error {
	return nil
}

// Run 执行任务
func (t *BaseTask) Run(ctx context.Context) error {
	return nil
}

// AfterRun 任务执行后
func (t *BaseTask) AfterRun(ctx context.Context) error {
	return nil
}
