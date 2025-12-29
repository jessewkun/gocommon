# response 模块

`response` 模块提供了一套统一的 API 返回结果封装，旨在标准化 Gin 框架应用的响应格式，并简化错误处理逻辑。

## 核心功能

-   **统一 API 响应结构**：定义标准化的 `APIResult` 数据结构。
-   **便捷的成功响应**：提供 `Success` 和 `SuccessWs` 函数。
-   **智能的错误处理**：`Error` 和 `ErrorWs` 函数能够识别 `common.CustomError`，并将其转换为标准响应，同时支持包装普通 `error`。
-   **灵活的自定义响应**：`Custom` 函数允许完全自定义返回码和消息。
-   **常用系统错误封装**：预定义了一系列常用的系统级错误，便于快速响应。

## API 返回结构：`APIResult`

所有 API 响应都将封装在这个结构中。

```go
type APIResult struct {
	Code    int         `json:"code"`     // 接口错误码，0表示成功，非0表示失败
	Message string      `json:"message"`  // 提示信息
	Data    interface{} `json:"data"`     // 返回数据
	TraceID string      `json:"trace_id"` // 请求唯一标识
}
```

## 核心函数

### 1. `Success(c *gin.Context, data interface{})`

用于返回成功的 API 响应。HTTP 状态码始终为 `200 OK`，JSON `code` 为 `0`。

**使用示例：**

```go
package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/response"
)

func main() {
	r := gin.Default()
	r.GET("/api/data", func(c *gin.Context) {
		result := map[string]string{"key": "value", "status": "ok"}
		response.Success(c, result)
	})
	r.Run(":8080")
}
```
**响应示例：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "key": "value",
        "status": "ok"
    },
    "trace_id": "..."
}
```

### 2. `Error(c *gin.Context, err error)`

用于返回错误的 API 响应。HTTP 状态码始终为 `200 OK`，JSON `code` 和 `message` 从 `err` 中提取。
该函数会尝试将 `err` 转换为 `common.CustomError`。如果成功，则使用 `CustomError` 中的 `Code` 和 `Error()` 作为响应；否则，会将 `err` 包装成一个默认的系统错误（`DefaultErrorCode`）进行响应。

**使用示例：**

```go
package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/response"
)

// 定义业务错误
var ErrUserNotFound = common.NewBusinessError(10001, errors.New("用户不存在"))

func main() {
	r := gin.Default()
	r.GET("/api/user/:id", func(c *gin.Context) {
		// 模拟业务逻辑错误
		userID := c.Param("id")
		if userID == "123" {
			response.Error(c, ErrUserNotFound) // 返回业务错误
			return
		}

		// 模拟一个普通Go错误
		if userID == "456" {
			response.Error(c, errors.New("未知内部错误")) // 会被包装为 DefaultErrorCode
			return
		}

		response.Success(c, gin.H{"id": userID, "name": "Test User"})
	})
	r.Run(":8080")
}
```
**响应示例 (业务错误)：**
```json
{
    "code": 10001,
    "message": "用户不存在",
    "data": {},
    "trace_id": "..."
}
```
**响应示例 (普通错误)：**
```json
{
    "code": 10000,
    "message": "未知内部错误",
    "data": {},
    "trace_id": "..."
}
```

### 3. `Custom(c *gin.Context, code int, message string, data interface{})`

用于返回完全自定义 `code` 和 `message` 的 API 响应。

**使用示例：**

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/response"
)

func main() {
	r := gin.Default()
	r.GET("/api/custom", func(c *gin.Context) {
		response.Custom(c, 20001, "自定义成功消息", gin.H{"status": "ok"})
	})
	r.Run(":8080")
}
```

### 4. 预定义系统错误函数

`response` 模块预定义了一系列常用的系统级错误，可以直接调用，简化了错误处理。这些错误都使用 `common.NewSystemError` 创建，错误码小于 `10001`。

-   `SystemError(c *gin.Context)`: 系统错误 (Code: `1000`)
-   `ParamError(c *gin.Context)`: 参数错误 (Code: `1001`)
-   `ForbiddenError(c *gin.Context)`: 权限错误 (Code: `1002`)
-   `NotfoundError(c *gin.Context)`: 未找到资源 (Code: `1003`)
-   `RateLimiterError(c *gin.Context)`: 限流错误 (Code: `1004`)
-   `UnknownError(c *gin.Context)`: 未知错误 (Code: `1005`)

**使用示例：**

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/response"
)

func main() {
	r := gin.Default()
	r.GET("/api/admin", func(c *gin.Context) {
		// 假设没有权限
		response.ForbiddenError(c)
	})
	r.GET("/api/item/:id", func(c *gin.Context) {
		itemID := c.Param("id")
		if itemID == "invalid" {
			response.ParamError(c)
			return
		}
		if itemID == "404" {
			response.NotfoundError(c)
			return
		}
		response.Success(c, gin.H{"item": itemID})
	})
	r.Run(":8080")
}
```

## 注意事项与最佳实践

-   **HTTP 状态码**: 当前设计中，所有 API 响应（包括成功和失败）均返回 `http.StatusOK` (200)，实际业务状态码通过 `APIResult.Code` 字段传递。
-   **错误码规范**: 遵循 `common` 模块定义的错误码规范：业务错误码 (`≥ 10001`)，系统错误码 (`< 10001`)。
-   **WebSocket 响应**: `SuccessWs` 和 `ErrorWs` 用于 WebSocket 场景，它们返回 `[]byte` 和 `error`，不直接写入 `gin.Context`，以便与 WebSocket 协议兼容。
-   **错误判断**: 在业务逻辑中判断错误时，推荐使用 `common.IsCode(err, code)` 而不是直接类型断言并比较 `Code` 字段，以支持更健壮的错误链判断。
