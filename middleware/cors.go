package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type CrosConfig struct {
	AllowedOrigins map[string]bool
	AllowMethods   []string
	AllowHeaders   []string
}

// 配置 CORS
func Cros(crosConfig CrosConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:     crosConfig.AllowMethods,
		AllowHeaders:     crosConfig.AllowHeaders,
		AllowCredentials: true,
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool {
			return crosConfig.AllowedOrigins[origin]
		},
	})
}
