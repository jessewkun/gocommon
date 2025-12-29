# safego 模块

`safego` 模块提供了一套工具，用于安全地执行代码块和管理 goroutine，能自动捕获并记录 panic，防止单个任务的崩溃导致整个服务宕机。

## 核心组件

### 1. `SafeGo(ctx context.Context, fun func())`

`SafeGo` 函数用于在**当前 goroutine** 中安全地执行一个函数。它本身**不会**启动新的 goroutine。

它的主要职责是为传入的函数包裹一个 `defer-recover` 块。一旦函数执行过程中发生 panic，`SafeGo` 会捕获它，记录详细的错误和堆栈信息，并允许当前 goroutine 继续正常执行，避免程序崩溃。

**使用场景：** 在一个已有的 goroutine 中，保护一个可能会 panic 的危险操作。

**示例：**

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/jessewkun/gocommon/safego"
)

func main() {
    ctx := context.Background()

    fmt.Println("准备启动一个带 panic 保护的 goroutine...")

    go func() {
        // 在 goroutine 内部使用 SafeGo 保护任务
        safego.SafeGo(ctx, func() {
            fmt.Println("任务开始执行...")
            time.Sleep(1 * time.Second)
            // 模拟一个 panic
            panic("something went wrong!")
        })

        // 由于 panic 被 SafeGo 捕获，这里的代码依然会执行
        fmt.Println("这个 goroutine 的后续任务...")
    }()

    // 等待 goroutine 执行，观察程序是否崩溃
    time.Sleep(3 * time.Second)
    fmt.Println("主程序正常结束。")
}
```

### 2. `WaitGroupWrapper`

`WaitGroupWrapper` 是 `sync.WaitGroup` 的一层封装，旨在简化“安全地启动并等待一组 goroutine”的常见模式。

#### 方法: `wg.Wrap(ctx context.Context, f func())`

`Wrap` 方法会启动一个**新的、受 panic 保护的 goroutine** 来执行传入的函数 `f`。它会自动处理 `wg.Add(1)` 和 `defer wg.Done()` 的逻辑。

**使用场景：** 当你需要并行执行多个可能失败的任务，并等待它们全部完成后再继续主流程时。

**示例：**

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/jessewkun/gocommon/safego"
)

func main() {
    var wg safego.WaitGroupWrapper
    ctx := context.Background()

    fmt.Println("使用 Wrap 启动 3 个并发任务...")

    for i := 1; i <= 3; i++ {
        taskID := i
        wg.Wrap(ctx, func() {
            fmt.Printf("任务 #%d 正在执行...\n", taskID)
            time.Sleep(time.Duration(taskID) * 500 * time.Millisecond)
            if taskID == 2 {
                // 模拟任务 #2 发生 panic
                panic(fmt.Sprintf("任务 #%d 失败!", taskID))
            }
            fmt.Printf("任务 #%d 完成。\n", taskID)
        })
    }

    // 等待所有通过 Wrap 启动的 goroutine 结束
    // 即使任务 #2 panic 了，wg.Wait() 也能正常工作
    wg.Wait()

    fmt.Println("所有并发任务已执行完毕。")
}
```
