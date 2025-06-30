package middleware

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"
)

// JWTConfig JWT配置
type JWTConfig struct {
	// JWT密钥
	SecretKey []byte
	// 过期时间
	Expiration time.Duration
	// 刷新时间
	RefreshTime time.Duration
	// 是否启用黑名单
	EnableBlacklist bool
	// 黑名单过期时间
	BlacklistExpiration time.Duration
	// 是否记录日志
	EnableLog bool
	// 自定义错误处理函数
	ErrorHandler func(c *gin.Context, err error)
}

// DefaultJWTConfig 返回默认配置
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:           []byte("your-256-bit-secret"),
		Expiration:          time.Hour * 24,     // 24小时
		RefreshTime:         time.Hour * 24 * 7, // 7天
		EnableBlacklist:     true,
		BlacklistExpiration: time.Hour * 24 * 7, // 7天
		EnableLog:           true,
		ErrorHandler:        nil,
	}
}

var (
	jwtConfig      *JWTConfig
	jwtConfigOnce  sync.Once
	blacklist      = make(map[string]time.Time)
	blacklistMutex sync.RWMutex
)

// Claims JWT声明
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// JwtAuth 返回一个JWT认证中间件
func JwtAuth(config *JWTConfig) gin.HandlerFunc {
	jwtConfigOnce.Do(func() {
		if config == nil {
			config = DefaultJWTConfig()
		}
		jwtConfig = config

		// 启动黑名单清理
		if config.EnableBlacklist {
			go func() {
				for {
					time.Sleep(time.Hour)
					cleanupBlacklist()
				}
			}()
		}
	})

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handleError(c, "Missing authorization header")
			return
		}

		// 检查 Authorization header 格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			handleError(c, "Invalid authorization header format")
			return
		}

		// 检查黑名单
		if jwtConfig.EnableBlacklist {
			blacklistMutex.RLock()
			if _, exists := blacklist[parts[1]]; exists {
				blacklistMutex.RUnlock()
				handleError(c, "Token has been revoked")
				return
			}
			blacklistMutex.RUnlock()
		}

		// 解析JWT
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return jwtConfig.SecretKey, nil
		})

		if err != nil {
			handleError(c, "Unauthorized: "+err.Error())
			return
		}

		if !token.Valid {
			handleError(c, "Invalid token")
			return
		}

		// 检查是否需要刷新token
		if time.Until(claims.ExpiresAt.Time) < jwtConfig.RefreshTime {
			newToken, err := refreshToken(claims)
			if err != nil {
				if jwtConfig.EnableLog {
					logger.Error(c.Request.Context(), TAG, err)
				}
			} else {
				c.Header("X-New-Token", newToken)
			}
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// CreateJwtToken 创建JWT token
func CreateJwtToken(userID int) (string, error) {
	expirationTime := time.Now().Add(jwtConfig.Expiration)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtConfig.SecretKey)
}

// RevokeToken 撤销token
func RevokeToken(token string) {
	if !jwtConfig.EnableBlacklist {
		return
	}

	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

	blacklist[token] = time.Now().Add(jwtConfig.BlacklistExpiration)
}

// refreshToken 刷新token
func refreshToken(claims *Claims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(jwtConfig.Expiration))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtConfig.SecretKey)
}

// cleanupBlacklist 清理过期的黑名单记录
func cleanupBlacklist() {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

	now := time.Now()
	for token, expireTime := range blacklist {
		if now.After(expireTime) {
			delete(blacklist, token)
		}
	}
}

// handleError 处理错误
func handleError(c *gin.Context, errMsg string) {
	if jwtConfig.EnableLog {
		logger.Warn(c.Request.Context(), TAG, "JWT error: %s", errMsg)
	}

	if jwtConfig.ErrorHandler != nil {
		jwtConfig.ErrorHandler(c, errors.New(errMsg))
	} else {
		c.JSON(http.StatusUnauthorized, response.ErrorResp(c, errors.New(errMsg)))
	}
	c.Abort()
}
