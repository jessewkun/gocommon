package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// 是否记录日志
	EnableLog bool
	// 自定义错误处理函数
	ErrorHandler func(c *gin.Context, err error)
	// 自定义成功处理函数
	SuccessHandler func(c *gin.Context, userID int)
	// 是否在上下文中设置用户ID
	SetUserID bool
	// 用户ID的上下文键
	UserIDKey string
}

// DefaultAuthConfig 返回默认配置
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		EnableLog:      true,
		ErrorHandler:   nil,
		SuccessHandler: nil,
		SetUserID:      true,
		UserIDKey:      "user_id",
	}
}

// CheckLogin 检查登录状态
func CheckLogin(checkFunc func(c *gin.Context) (int, error), config *AuthConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return func(c *gin.Context) {
		userID, err := checkFunc(c)
		if err != nil {
			if config.EnableLog {
				logger.Warn(c.Request.Context(), TAGNAME, "Login check failed: %v", err)
			}

			if config.ErrorHandler != nil {
				config.ErrorHandler(c, err)
			} else {
				c.JSON(http.StatusUnauthorized, response.ErrorResp(c, err))
			}
			c.Abort()
			return
		}

		if config.SetUserID {
			c.Set(config.UserIDKey, userID)
		}

		if config.SuccessHandler != nil {
			config.SuccessHandler(c, userID)
		}

		c.Next()
	}
}

// NeedLogin 需要登录
func NeedLogin(checkFunc func(c *gin.Context) (int, error), config *AuthConfig) gin.HandlerFunc {
	return CheckLogin(checkFunc, config)
}
