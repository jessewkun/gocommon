package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CrosConfig 配置 CORS
type CrosConfig struct {
	AllowOrigins               []string `mapstructure:"allow_origins" json:"allow_origins"`
	AllowMethods               []string `mapstructure:"allow_methods" json:"allow_methods"`
	AllowHeaders               []string `mapstructure:"allow_headers" json:"allow_headers"`
	AllowOriginWithContextFunc func(c *gin.Context, origin string) bool
}

// Cros 配置 CORS
func Cros(crosConfig CrosConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:               crosConfig.AllowMethods,
		AllowHeaders:               crosConfig.AllowHeaders,
		AllowOrigins:               crosConfig.AllowOrigins,
		AllowCredentials:           true,
		AllowOriginWithContextFunc: crosConfig.AllowOriginWithContextFunc,
	})
}
