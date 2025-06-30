package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.GlobalLimiter)
	assert.NotNil(t, config.IPLimiters)
	assert.NotNil(t, config.IPLastUsed)
	assert.Equal(t, rate.Limit(10), config.IPLimit)
	assert.Equal(t, 20, config.IPBurst)
	assert.True(t, config.EnableIPLimit)
	assert.True(t, config.EnableLog)
	assert.Equal(t, time.Minute*10, config.CleanupInterval)
	assert.Equal(t, time.Minute*10, config.IPExpiration)
	assert.Empty(t, config.Whitelist)
	assert.False(t, config.WhitelistSkipGlobal)
	assert.Nil(t, config.WhitelistChecker)
}

func TestRateLimiter_BasicFunctionality(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(10), 2), // Burst=2
		IPLimit:         rate.Limit(5),
		IPBurst:         2,
		EnableIPLimit:   true,
		EnableLog:       false,
		CleanupInterval: time.Minute * 10,
		IPExpiration:    time.Minute * 10,
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("正常请求应该通过", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("IP限流应该生效", func(t *testing.T) {
		limitCount := 0
		for i := 0; i < 50; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			router.ServeHTTP(w, req)
			if strings.Contains(w.Body.String(), "too many requests") {
				limitCount++
			}
		}
		assert.Greater(t, limitCount, 0, "应该有请求被限流")
	})
}

func TestRateLimiter_Whitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &RateLimiterConfig{
		GlobalLimiter:       rate.NewLimiter(rate.Limit(1), 2),
		IPLimit:             rate.Limit(1),
		IPBurst:             2,
		EnableIPLimit:       true,
		EnableLog:           false,
		Whitelist:           []string{"127.0.0.1", "192.168.1.100"},
		WhitelistSkipGlobal: true,
		CleanupInterval:     time.Minute * 10,
		IPExpiration:        time.Minute * 10,
		IPLimiters:          make(map[string]*rate.Limiter),
		IPLastUsed:          make(map[string]time.Time),
		stopCleanup:         make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("白名单IP应该不受限流", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "白名单IP应该不受限流")
			assert.Contains(t, w.Body.String(), "success")
		}
	})

	t.Run("非白名单IP应该受限流", func(t *testing.T) {
		limitCount := 0
		for i := 0; i < 50; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.200:12345"
			router.ServeHTTP(w, req)
			if strings.Contains(w.Body.String(), "too many requests") {
				limitCount++
			}
		}
		assert.Greater(t, limitCount, 0, "非白名单IP应该被限流")
	})
}

func TestRateLimiter_CustomWhitelistChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &RateLimiterConfig{
		GlobalLimiter: rate.NewLimiter(rate.Limit(1), 2),
		IPLimit:       rate.Limit(1),
		IPBurst:       2,
		EnableIPLimit: true,
		EnableLog:     false,
		WhitelistChecker: func(ip string) bool {
			return strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.")
		},
		WhitelistSkipGlobal: true,
		CleanupInterval:     time.Minute * 10,
		IPExpiration:        time.Minute * 10,
		IPLimiters:          make(map[string]*rate.Limiter),
		IPLastUsed:          make(map[string]time.Time),
		stopCleanup:         make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("自定义白名单检查器应该生效", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "内网IP应该不受限流")
			assert.Contains(t, w.Body.String(), "success")
		}
		limitCount := 0
		for i := 0; i < 50; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "8.8.8.8:12345"
			router.ServeHTTP(w, req)
			if strings.Contains(w.Body.String(), "too many requests") {
				limitCount++
			}
		}
		assert.Greater(t, limitCount, 0, "外网IP应该被限流")
	})
}

func TestRateLimiter_GlobalLimiterOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(5), 2),
		EnableIPLimit:   false,
		EnableLog:       false,
		IPBurst:         2,
		CleanupInterval: time.Minute * 10,
		IPExpiration:    time.Minute * 10,
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("只有全局限流时应该正常工作", func(t *testing.T) {
		limitCount := 0
		for i := 0; i < 50; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			router.ServeHTTP(w, req)
			if strings.Contains(w.Body.String(), "too many requests") {
				limitCount++
			}
		}
		assert.Greater(t, limitCount, 0, "应该有请求被全局限流")
	})
}

func TestRateLimiter_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimiter(nil)) // 使用默认配置

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("nil配置应该使用默认配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(100), 200),
		IPLimit:         rate.Limit(20),
		IPBurst:         40,
		EnableIPLimit:   true,
		EnableLog:       false,
		CleanupInterval: time.Minute * 10, // 添加清理间隔
		IPExpiration:    time.Minute * 10, // 添加过期时间
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("并发请求应该正确处理", func(t *testing.T) {
		const numGoroutines = 10
		const requestsPerGoroutine = 10

		var wg sync.WaitGroup
		successCount := int32(0)
		limitCount := int32(0)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < requestsPerGoroutine; j++ {
					w := httptest.NewRecorder()
					req, _ := http.NewRequest("GET", "/test", nil)
					req.RemoteAddr = "192.168.1.1:12345"

					router.ServeHTTP(w, req)

					if w.Code == http.StatusOK {
						atomic.AddInt32(&successCount, 1)
					} else {
						atomic.AddInt32(&limitCount, 1)
					}
				}
			}(i)
		}

		wg.Wait()

		totalRequests := numGoroutines * requestsPerGoroutine
		assert.Equal(t, int32(totalRequests), successCount+limitCount)
		assert.Greater(t, successCount, int32(0), "应该有成功的请求")
	})
}

func TestRateLimiter_ManagementFunctions(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(10), 20),
		IPLimit:         rate.Limit(5),
		IPBurst:         10,
		EnableIPLimit:   true,
		EnableLog:       false,
		CleanupInterval: time.Minute * 10, // 添加清理间隔
		IPExpiration:    time.Minute * 10, // 添加过期时间
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	t.Run("管理函数应该正常工作", func(t *testing.T) {
		// 先创建一些IP限流器
		getIPLimiter(config, "192.168.1.1")
		getIPLimiter(config, "192.168.1.2")
		getIPLimiter(config, "192.168.1.3")

		// 测试获取数量
		count := GetIPLimitersCount(config)
		assert.Equal(t, 3, count)

		// 测试获取信息
		info := GetIPLimitersInfo(config)
		assert.Len(t, info, 3)
		assert.Contains(t, info, "192.168.1.1")
		assert.Contains(t, info, "192.168.1.2")
		assert.Contains(t, info, "192.168.1.3")

		// 测试删除特定IP
		success := RemoveIPLimiter(config, "192.168.1.2")
		assert.True(t, success)

		// 验证删除结果
		count = GetIPLimitersCount(config)
		assert.Equal(t, 2, count)

		info = GetIPLimitersInfo(config)
		assert.Len(t, info, 2)
		assert.Contains(t, info, "192.168.1.1")
		assert.NotContains(t, info, "192.168.1.2")
		assert.Contains(t, info, "192.168.1.3")

		// 测试删除不存在的IP
		success = RemoveIPLimiter(config, "192.168.1.999")
		assert.False(t, success)
	})
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(10), 20),
		IPLimit:         rate.Limit(5),
		IPBurst:         10,
		EnableIPLimit:   true,
		EnableLog:       false,
		CleanupInterval: time.Millisecond * 100, // 快速清理用于测试
		IPExpiration:    time.Millisecond * 50,  // 快速过期
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	t.Run("清理功能应该正常工作", func(t *testing.T) {
		// 创建一些IP限流器
		getIPLimiter(config, "192.168.1.1")
		getIPLimiter(config, "192.168.1.2")

		// 验证初始状态
		assert.Equal(t, 2, GetIPLimitersCount(config))

		// 等待清理
		time.Sleep(time.Millisecond * 200)

		// 手动触发清理
		cleanupIPLimiters(config)

		// 验证清理结果
		count := GetIPLimitersCount(config)
		assert.Equal(t, 0, count, "过期的IP限流器应该被清理")
	})
}

func TestRateLimiter_StopCleanup(t *testing.T) {
	config := &RateLimiterConfig{
		GlobalLimiter:   rate.NewLimiter(rate.Limit(10), 20),
		IPLimit:         rate.Limit(5),
		IPBurst:         10,
		EnableIPLimit:   true,
		EnableLog:       false,
		CleanupInterval: time.Millisecond * 50,
		IPExpiration:    time.Millisecond * 25,
		IPLimiters:      make(map[string]*rate.Limiter),
		IPLastUsed:      make(map[string]time.Time),
		stopCleanup:     make(chan struct{}),
	}

	t.Run("停止清理功能应该正常工作", func(t *testing.T) {
		// 启动限流器（会启动清理goroutine）
		router := gin.New()
		router.Use(RateLimiter(config))

		// 等待一下让清理goroutine启动
		time.Sleep(time.Millisecond * 10)

		// 停止清理
		StopCleanup(config)

		// 验证停止信号已发送
		select {
		case <-config.stopCleanup:
			// 正常情况
		default:
			t.Error("停止信号应该已发送")
		}
	})
}

func TestRateLimiter_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		GlobalLimiter: rate.NewLimiter(rate.Limit(1), 1), // 很严格的限制
		EnableIPLimit: false,
		EnableLog:     false,
		IPLimiters:    make(map[string]*rate.Limiter),
		IPLastUsed:    make(map[string]time.Time),
		stopCleanup:   make(chan struct{}),
	}

	router := gin.New()
	router.Use(RateLimiter(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("限流响应格式应该正确", func(t *testing.T) {
		// 第一个请求应该成功
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 第二个请求应该被限流
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)                   // 限流响应也是200状态码
		assert.Contains(t, w2.Body.String(), "too many requests") // 检查错误信息
	})
}
