// Package middleware 提供中间件功能
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/response"
)

// CheckLoginFunc 自定义登录态检查签名
//
// 下边的 CheckLogin 中间件仅仅是一个容器，具体的登录态检查逻辑需要在业务中实现该签名
// 登录态在不同的业务中可能保存的信息不同，比如有的业务保存的是用户id，有的业务保存的是用户名，有的业务保存的是用户信息
// 强烈建议保存 user_id 到 gin.Context 中，方便后续使用
type CheckLoginFunc func(c *gin.Context) error

// NeedLoginFunc 下边的 NeedLogin 中间件仅仅是一个容器，具体的登录态检查逻辑需要在业务中实现该签名
// 同上
type NeedLoginFunc func(c *gin.Context)

// CheckLogin 检查登录态
func CheckLogin(fun CheckLoginFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fun(c); err != nil {
			response.Error(c, common.NewCustomError(10001, err))
			c.Abort()
			return
		}
		c.Next()
	}
}

// NeedLogin 需要登录态
func NeedLogin(fun NeedLoginFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fun(c)
		c.Next()
	}
}
