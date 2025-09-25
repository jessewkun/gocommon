// Package cron 提供定时任务管理功能
package cron

import (
	"context"
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
	// Name 任务名称
	Name() string
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

// 全局管理器实例，用于自动注册
var globalManager *Manager

// 延迟注册的任务队列
var pendingTasks []Task

// NewManager 创建定时任务管理器
func NewManager() *Manager {
	manager := &Manager{
		cron:  cron.New(cron.WithSeconds()),
		tasks: make(map[string]Task),
	}
	globalManager = manager

	// 注册所有延迟注册的任务
	for _, task := range pendingTasks {
		manager.RegisterTask(task)
	}
	pendingTasks = nil

	return manager
}

// AutoRegisterTask 自动注册任务到全局管理器
func AutoRegisterTask(task Task) error {
	if globalManager != nil {
		// 管理器已创建，直接注册
		return globalManager.RegisterTask(task)
	} else {
		// 管理器未创建，加入延迟注册队列
		pendingTasks = append(pendingTasks, task)
		return nil
	}
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

	name := task.Name()
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	if _, exists := m.tasks[name]; exists {
		return fmt.Errorf("task %s already registered", name)
	}

	m.tasks[name] = task
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
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	ctx := m.cron.Stop()
	<-ctx.Done()
	m.running = false

	logger.Info(context.Background(), "CRON", "Stopped cron manager")
}

// runTask 执行单个任务
func (m *Manager) runTask(name string, task Task) error {
	startTime := time.Now()

	taskCtx := context.Background()
	taskCtx = context.WithValue(taskCtx, constant.CtxTraceID, uuid.New().String())

	logger.Info(taskCtx, "CRON", "Task %s started", name)

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
				resultChan <- result{err: err}
				return
			}

			// 执行主任务逻辑
			err = task.Run(taskCtx)

			// 执行 AfterRun 钩子（无论主任务成功还是失败）
			if afterErr := task.AfterRun(taskCtx); afterErr != nil {
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

	var err error
	timeout := task.Timeout()

	if timeout > 0 {
		select {
		case res := <-resultChan:
			err = res.err
		case <-time.After(timeout):
			err = fmt.Errorf("task %s timeout after %v", name, timeout)
		}
	} else {
		res := <-resultChan
		err = res.err
	}

	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorWithMsg(taskCtx, "CRON", "Task %s failed after %v: %v", name, duration, err)
	} else {
		logger.Info(taskCtx, "CRON", "Task %s completed successfully in %v", name, duration)
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
// 参数：
//   - ctx: 上下文，用于链路追踪和超时控制
//   - taskName: 任务名称
//
// 返回：
//   - error: 执行过程中的错误
func (m *Manager) RunTask(ctx context.Context, taskName string) error {
	m.mu.RLock()
	task, exists := m.tasks[taskName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", taskName)
	}

	if !task.Enabled() {
		return fmt.Errorf("task %s is disabled", taskName)
	}

	return m.runTask(taskName, task)
}
