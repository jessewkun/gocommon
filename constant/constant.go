// Package constant 定义项目中使用的常量
package constant

// ContextKey 定义 context key 的类型
type ContextKey string

// 用于设置gin.Context中的key来存储返回结果，并在中间件中获取记录日志

// CtxAPIOutput 用于设置gin.Context中的key来存储返回结果，并在中间件中获取记录日志
const CtxAPIOutput ContextKey = "api_output"

// CtxTraceID 用于设置gin.Context中的key来存储trace_id
const CtxTraceID ContextKey = "trace_id"

// CtxUserID 用于设置gin.Context中的key来存储user_id
const CtxUserID ContextKey = "user_id"

// CtxCurrentRequestPath 用于设置gin.Context中的key来存储当前请求路径
const CtxCurrentRequestPath ContextKey = "current_request_path"

// CtxIsPts 用于设置gin.Context中的key来存储是否是压测
const CtxIsPts ContextKey = "is_pts"
