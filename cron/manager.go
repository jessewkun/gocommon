// Package cron 提供定时任务管理功能
package cron

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jessewkun/gocommon/constant"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/safego"
	"github.com/robfig/cron/v3"
)

// Task 定时任务接口
type Task interface {
	// Key 任务标识
	Key() string
	// Spec 任务调度表达式，支持 cron 格式
	Spec() string
	// BeforeRun 任务执行前
	BeforeRun(ctx context.Context) error
	// AfterRun 任务执行后
	AfterRun(ctx context.Context) error
	// Run 执行任务
	Run(ctx context.Context) error
	// Timeout 任务超时时间，0表示不超时
	Timeout() time.Duration
	// Enabled 任务是否启用
	Enabled() bool
}

// Manager 定时任务管理器
type Manager struct {
	cron    *cron.Cron
	tasks   map[string]Task
	mu      sync.RWMutex
	running bool
}

// NewManager 创建定时任务管理器
func NewManager() *Manager {
	manager := &Manager{
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
		),
		tasks: make(map[string]Task),
	}
	return manager
}

// RegisterTask 注册定时任务
func (m *Manager) RegisterTask(task Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("cannot register task after manager started")
	}

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	key := task.Key()
	if key == "" {
		return fmt.Errorf("task key cannot be empty")
	}

	if _, exists := m.tasks[key]; exists {
		return fmt.Errorf("task %s already registered", key)
	}

	m.tasks[key] = task
	return nil
}

// Start 启动定时任务管理器
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("manager already started")
	}

	for name, task := range m.tasks {
		if !task.Enabled() {
			logger.Info(ctx, "CRON", "Skipping disabled task: %s", name)
			continue
		}

		_, err := m.cron.AddFunc(task.Spec(), func() {
			runErr := m.runTask(name, task)
			if runErr != nil {
				logger.ErrorWithMsg(ctx, "CRON", "Task %s failed: %v", name, runErr)
			}
		})
		if err != nil {
			return fmt.Errorf("failed to add task %s: %w", name, err)
		}
		logger.Info(ctx, "CRON", "Registered task: %s", name)
	}

	m.cron.Start()
	m.running = true

	logger.Info(ctx, "CRON", "Started cron manager with %d tasks", len(m.tasks))
	return nil
}

// Stop 停止定时任务管理器
func (m *Manager) Stop(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	c := m.cron.Stop()
	<-c.Done()
	m.running = false

	logger.Info(ctx, "CRON", "Stopped cron manager")
}

// runTask 执行单个任务
func (m *Manager) runTask(name string, task Task) error {
	startTime := time.Now()

	// Create a base context with trace ID
	baseCtx := context.Background()
	baseCtx = context.WithValue(baseCtx, constant.CtxTraceID, uuid.New().String())

	logger.Info(baseCtx, "CRON", "Task %s started", name)

	// Prepare the final context for the task
	var (
		taskCtx context.Context
		cancel  context.CancelFunc
	)

	timeout := task.Timeout()
	if timeout > 0 {
		taskCtx, cancel = context.WithTimeout(baseCtx, timeout)
	} else {
		taskCtx, cancel = context.WithCancel(baseCtx)
	}
	defer cancel() // Ensure cancel is always called to free resources

	type result struct {
		err error
	}
	resultChan := make(chan result, 1)

	go func() {
		var err error
		safego.SafeGo(taskCtx, func() {
			// 执行 BeforeRun 钩子
			if beforeErr := task.BeforeRun(taskCtx); beforeErr != nil {
				err = fmt.Errorf("before run failed: %w", beforeErr)
				return
			}

			// 检查上下文是否在任务开始前就已超时
			if taskCtx.Err() != nil {
				err = taskCtx.Err()
				return
			}

			// 执行主任务逻辑
			err = task.Run(taskCtx)

			// AfterRun 钩子应该总是被执行，即使任务超时。
			// 我们使用 baseCtx 来运行它，以避免它因为 taskCtx 被取消而无法执行。
			if afterErr := task.AfterRun(baseCtx); afterErr != nil {
				if err != nil {
					// 如果主任务也失败了，记录两个错误
					err = fmt.Errorf("main task failed: %w, after run also failed: %w", err, afterErr)
				} else {
					// 如果主任务成功但 AfterRun 失败
					err = fmt.Errorf("after run failed: %w", afterErr)
				}
			}
		})
		resultChan <- result{err: err}
	}()

	// 等待任务执行结果或超时/取消
	var err error
	select {
	case res := <-resultChan:
		err = res.err
	case <-taskCtx.Done():
		err = taskCtx.Err()
	}

	// 如果错误是由于上下文超时/取消引起的，包装成更明确的超时错误信息
	if errors.Is(err, context.DeadlineExceeded) {
		err = fmt.Errorf("task %s timeout after %v", name, timeout)
	}

	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorWithMsg(baseCtx, "CRON", "Task %s failed after %v: %v", name, duration, err)
	} else {
		logger.Info(baseCtx, "CRON", "Task %s completed successfully in %v", name, duration)
	}
	return err
}

// GetTaskNames 获取所有已注册的任务名称
func (m *Manager) GetTaskNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tasks))
	for name := range m.tasks {
		names = append(names, name)
	}
	return names
}

// IsRunning 检查管理器是否正在运行
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// RunTask 手动执行指定的任务
func (m *Manager) RunTask(ctx context.Context, taskName string) error {
	m.mu.RLock()
	task, exists := m.tasks[taskName]
	m.mu.RUnlock()

	if !exists {
		// 注意：这里的任务未找到，可能是因为配置中未启用或未定义
		return fmt.Errorf("task %s not found (it may be disabled or not defined in config)", taskName)
	}

	// 直接执行，因为 m.tasks 中存储的已经是被 ConfigurableTask 包装过的任务
	// 它自带了从配置中读取的超时等信息
	return m.runTask(taskName, task)
}
