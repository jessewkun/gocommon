// Package cron 提供定时任务的具体实现
package cron

import (
	"context"
	"time"
)

// ExampleTask 示例任务
type ExampleTask struct {
	BaseTask
}

// BeforeRun 任务执行前
func (t *ExampleTask) BeforeRun(ctx context.Context) error {
	// 在这里可以做一些准备工作，比如：
	// - 检查依赖服务是否可用
	// - 初始化资源
	// - 记录任务开始状态
	return nil
}

// AfterRun 任务执行后
func (t *ExampleTask) AfterRun(ctx context.Context) error {
	// 在这里可以做一些清理工作，比如：
	// - 释放资源
	// - 更新任务状态
	// - 发送通知
	return nil
}

// Run 执行任务
func (t *ExampleTask) Run(ctx context.Context) error {
	// 模拟业务逻辑
	time.Sleep(5 * time.Second)
	return nil
}

// init 初始化示例任务
// func init() {
// 	AutoRegisterTask(&ExampleTask{
// 		BaseTask: BaseTask{
// 			TaskName:    "example_task",
// 			TaskEnabled: true,
// 			TaskSpec:    "0 * * * * *", // 每分钟执行一次
// 			TaskTimeout: 2 * time.Second,
// 		},
// 	})
// }

// 手动执行任务的示例用法：
//
// // 1. 同步执行任务
// func runTaskSync() {
//     ctx := context.Background()
//     err := RunGlobalTask(ctx, "example_task")
//     if err != nil {
//         fmt.Printf("任务执行失败: %v\n", err)
//     } else {
//         fmt.Println("任务执行成功")
//     }
// }
//
// // 2. 异步执行任务
// func runTaskAsync() {
//     ctx := context.Background()
//     resultChan := RunGlobalTaskAsync(ctx, "example_task")
//
//     // 等待任务完成
//     err := <-resultChan
//     if err != nil {
//         fmt.Printf("任务执行失败: %v\n", err)
//     } else {
//         fmt.Println("任务执行成功")
//     }
// }
//
// // 3. 使用 Manager 实例执行任务
// func runTaskWithManager() {
//     manager := NewManager()
//
//     // 注册任务
//     manager.RegisterTask(&ExampleTask{
//         BaseTask: BaseTask{
//             TaskName:    "example_task",
//             TaskEnabled: true,
//             TaskSpec:    "0 * * * * *",
//             TaskTimeout: 2 * time.Second,
//         },
//     })
//
//     // 手动执行任务
//     ctx := context.Background()
//     err := manager.RunTask(ctx, "example_task")
//     if err != nil {
//         fmt.Printf("任务执行失败: %v\n", err)
//     }
// }
