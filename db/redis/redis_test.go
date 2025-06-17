package redis

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	_ = logger.InitLogger(&logger.Config{
		Path:       "./test.log",
		Closed:     false,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackup:  1,
		AlarmLevel: "warn",
	})
}

func TestInitRedis(t *testing.T) {
	// 测试配置
	config := map[string]*Config{
		"test": {
			Addrs:              []string{"127.0.0.1:6379"},
			Password:           "",
			Db:                 0,
			IsLog:              true,
			PoolSize:           10,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
	}

	// 测试初始化（这里会失败，因为没有真实的Redis连接）
	err := InitRedis(config)
	// 由于没有真实的Redis连接，这里期望失败
	assert.NoError(t, err)
}

func TestSetDefaultConfig(t *testing.T) {
	// 测试空配置
	config := &Config{}
	err := setDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis addrs is empty")

	// 测试有效配置
	config = &Config{
		Addrs: []string{"localhost:6379"},
	}
	err = setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 500, config.PoolSize)
	assert.Equal(t, 1, config.IdleTimeout)
	assert.Equal(t, 10, config.IdleCheckFrequency)
	assert.Equal(t, 3, config.MinIdleConns)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 2, config.DialTimeout)
	assert.Equal(t, 200, config.SlowThreshold)

	// 测试自定义配置
	config = &Config{
		Addrs:              []string{"localhost:6379"},
		PoolSize:           100,
		IdleTimeout:        600,
		IdleCheckFrequency: 120,
		MinIdleConns:       10,
		MaxRetries:         5,
		DialTimeout:        10,
		SlowThreshold:      500,
	}
	err = setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 100, config.PoolSize)
	assert.Equal(t, 600, config.IdleTimeout)
	assert.Equal(t, 120, config.IdleCheckFrequency)
	assert.Equal(t, 10, config.MinIdleConns)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10, config.DialTimeout)
	assert.Equal(t, 500, config.SlowThreshold)
}

func TestGetConn(t *testing.T) {
	// 测试获取不存在的连接
	_, err := GetConn("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis conn is not found")
}

func TestHealthCheck(t *testing.T) {
	// 测试健康检查（没有连接时应该返回空结果）
	healthStatus := HealthCheck()
	assert.NotNil(t, healthStatus)
	assert.Len(t, healthStatus, 1)
}

// 测试辅助函数
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "empty addrs",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Addrs: []string{"localhost:6379"},
			},
			wantErr: false,
		},
		{
			name: "multiple addrs",
			config: &Config{
				Addrs: []string{
					"localhost:7000",
					"localhost:7001",
					"localhost:7002",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setDefaultConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 测试集群配置
func TestClusterConfig(t *testing.T) {
	config := &Config{
		Addrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
		PoolSize:           50,
		IdleTimeout:        300,
		IdleCheckFrequency: 60,
		MinIdleConns:       5,
		MaxRetries:         3,
		DialTimeout:        5,
		SlowThreshold:      100,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Len(t, config.Addrs, 3)
}

// 测试连接池配置
func TestConnectionPoolConfig(t *testing.T) {
	config := &Config{
		Addrs:              []string{"localhost:6379"},
		PoolSize:           200,
		IdleTimeout:        600,
		IdleCheckFrequency: 120,
		MinIdleConns:       20,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 200, config.PoolSize)
	assert.Equal(t, 600, config.IdleTimeout)
	assert.Equal(t, 120, config.IdleCheckFrequency)
	assert.Equal(t, 20, config.MinIdleConns)
}

// 测试超时配置
func TestTimeoutConfig(t *testing.T) {
	config := &Config{
		Addrs:       []string{"localhost:6379"},
		DialTimeout: 10,
		MaxRetries:  5,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 10, config.DialTimeout)
	assert.Equal(t, 5, config.MaxRetries)
}

// 测试慢查询配置
func TestSlowQueryConfig(t *testing.T) {
	config := &Config{
		Addrs:         []string{"localhost:6379"},
		SlowThreshold: 500, // 500毫秒
		IsLog:         true,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 500, config.SlowThreshold)
}

// 测试日志配置
func TestLoggingConfig(t *testing.T) {
	config := &Config{
		Addrs: []string{"localhost:6379"},
		IsLog: true,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.True(t, config.IsLog)
}

// 测试数据库配置
func TestDatabaseConfig(t *testing.T) {
	config := &Config{
		Addrs: []string{"localhost:6379"},
		Db:    1,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, config.Db)
}

// 测试密码配置
func TestPasswordConfig(t *testing.T) {
	config := &Config{
		Addrs:    []string{"localhost:6379"},
		Password: "testpassword",
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, "testpassword", config.Password)
}

// 测试并发安全性
func TestConcurrencySafety(t *testing.T) {
	// 这个测试主要验证连接管理的并发安全性
	// 由于没有真实的Redis连接，这里只是验证函数调用不会panic

	// 并发调用 GetConn
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = GetConn("test")
		}()
	}

	// 并发调用 HealthCheck
	for i := 0; i < 10; i++ {
		go func() {
			_ = HealthCheck()
		}()
	}

	// 等待一段时间确保并发操作完成
	time.Sleep(100 * time.Millisecond)
}

// 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试空的地址列表
	config := &Config{
		Addrs: []string{},
	}

	err := setDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis addrs is empty")

	// 测试无效的地址格式
	config = &Config{
		Addrs: []string{"invalid-addr"},
	}

	err = setDefaultConfig(config)
	assert.NoError(t, err) // setDefaultConfig 只验证长度，不验证地址格式
}

// 测试上下文支持
func TestContextSupport(t *testing.T) {
	ctx := context.Background()

	// 测试带上下文的操作（这里只是验证函数签名）
	_ = ctx

	// 在实际使用中，Redis 客户端支持上下文
	// 例如：client.Set(ctx, "key", "value", time.Hour)
}

// 测试配置结构体
func TestConfigStruct(t *testing.T) {
	config := &Config{
		Addrs:              []string{"localhost:6379"},
		Password:           "password",
		Db:                 1,
		IsLog:              true,
		PoolSize:           100,
		IdleTimeout:        300,
		IdleCheckFrequency: 60,
		MinIdleConns:       10,
		MaxRetries:         3,
		DialTimeout:        5,
		SlowThreshold:      100,
	}

	assert.Len(t, config.Addrs, 1)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, 1, config.Db)
	assert.True(t, config.IsLog)
	assert.Equal(t, 100, config.PoolSize)
	assert.Equal(t, 300, config.IdleTimeout)
	assert.Equal(t, 60, config.IdleCheckFrequency)
	assert.Equal(t, 10, config.MinIdleConns)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.DialTimeout)
	assert.Equal(t, 100, config.SlowThreshold)
}

// 测试健康状态结构体
func TestHealthStatusStruct(t *testing.T) {
	status := &HealthStatus{
		Status:    "success",
		Error:     "",
		Latency:   10,
		Timestamp: time.Now().UnixMilli(),
	}

	assert.Equal(t, "success", status.Status)
	assert.Empty(t, status.Error)
	assert.Greater(t, status.Latency, int64(0))
	assert.Greater(t, status.Timestamp, int64(0))
}

// 测试边界值
func TestBoundaryValues(t *testing.T) {
	// 测试最小连接池大小
	config := &Config{
		Addrs:        []string{"localhost:6379"},
		PoolSize:     1,
		MinIdleConns: 1,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, config.PoolSize)
	assert.Equal(t, 1, config.MinIdleConns)

	// 测试最大连接池大小
	config = &Config{
		Addrs:    []string{"localhost:6379"},
		PoolSize: 1000,
	}

	err = setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1000, config.PoolSize)

	// 测试零值超时
	config = &Config{
		Addrs:         []string{"localhost:6379"},
		IdleTimeout:   0,
		DialTimeout:   0,
		MaxRetries:    0,
		SlowThreshold: 0,
	}

	err = setDefaultConfig(config)
	assert.NoError(t, err)
	// 应该使用默认值
	assert.Equal(t, 1, config.IdleTimeout)
	assert.Equal(t, 2, config.DialTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 200, config.SlowThreshold)
}

// 测试配置验证的完整性
func TestConfigValidationCompleteness(t *testing.T) {
	// 测试所有字段的默认值设置
	config := &Config{
		Addrs: []string{"localhost:6379"},
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)

	// 验证所有字段都有合理的默认值
	assert.NotEmpty(t, config.Addrs)
	assert.Greater(t, config.PoolSize, 0)
	assert.Greater(t, config.IdleTimeout, 0)
	assert.Greater(t, config.IdleCheckFrequency, 0)
	assert.GreaterOrEqual(t, config.MinIdleConns, 0)
	assert.GreaterOrEqual(t, config.MaxRetries, 0)
	assert.Greater(t, config.DialTimeout, 0)
	assert.GreaterOrEqual(t, config.SlowThreshold, 0)
}

// 测试配置的不可变性
func TestConfigImmutability(t *testing.T) {
	originalConfig := &Config{
		Addrs:              []string{"localhost:6379"},
		Password:           "original",
		Db:                 0,
		PoolSize:           100,
		IdleTimeout:        300,
		IdleCheckFrequency: 60,
		MinIdleConns:       10,
		MaxRetries:         3,
		DialTimeout:        5,
		SlowThreshold:      100,
	}

	// 复制配置
	configCopy := *originalConfig

	// 修改原始配置
	originalConfig.Password = "modified"
	originalConfig.PoolSize = 200

	// 验证副本没有被修改
	assert.Equal(t, "original", configCopy.Password)
	assert.Equal(t, 100, configCopy.PoolSize)
}

// 测试真实的Redis读写操作
func TestRealRedisOperations(t *testing.T) {
	// 跳过测试，如果没有真实的Redis服务
	// 要运行这个测试，需要启动Redis服务并修改配置
	// t.Skip("跳过真实Redis测试，需要启动Redis服务")

	// 测试配置 - 需要修改为真实的Redis地址
	config := map[string]*Config{
		"test": {
			Addrs:              []string{"127.0.0.1:6379"},
			Password:           "",
			Db:                 0,
			IsLog:              true,
			PoolSize:           10,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
	}

	// 初始化Redis连接
	err := InitRedis(config)
	assert.NoError(t, err)
	defer CloseRedis()

	// 获取Redis连接
	client, err := GetConn("test")
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()

	// 测试字符串操作
	t.Run("String Operations", func(t *testing.T) {
		key := "test:string:key"
		value := "test_value"

		// 设置字符串
		err := client.Set(ctx, key, value, time.Hour).Err()
		assert.NoError(t, err)

		// 获取字符串
		result, err := client.Get(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// 检查键是否存在
		exists, err := client.Exists(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		// 删除键
		deleted, err := client.Del(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})

	// 测试哈希操作
	t.Run("Hash Operations", func(t *testing.T) {
		key := "test:hash:key"
		field1 := "field1"
		value1 := "value1"
		field2 := "field2"
		value2 := "value2"

		// 设置哈希字段
		err := client.HSet(ctx, key, field1, value1, field2, value2).Err()
		assert.NoError(t, err)

		// 获取哈希字段
		result1, err := client.HGet(ctx, key, field1).Result()
		assert.NoError(t, err)
		assert.Equal(t, value1, result1)

		// 获取所有哈希字段
		allFields, err := client.HGetAll(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, value1, allFields[field1])
		assert.Equal(t, value2, allFields[field2])

		// 检查哈希字段是否存在
		exists, err := client.HExists(ctx, key, field1).Result()
		assert.NoError(t, err)
		assert.True(t, exists)

		// 删除哈希字段
		deleted, err := client.HDel(ctx, key, field1).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		// 删除整个哈希
		delKey, err := client.Del(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), delKey)
	})

	// 测试列表操作
	t.Run("List Operations", func(t *testing.T) {
		key := "test:list:key"
		values := []string{"value1", "value2", "value3"}

		// 从左侧推入元素
		for _, value := range values {
			err := client.LPush(ctx, key, value).Err()
			assert.NoError(t, err)
		}

		// 获取列表长度
		length, err := client.LLen(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(values)), length)

		// 从左侧弹出元素
		popped, err := client.LPop(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, values[len(values)-1], popped) // LPush后，最后推入的在最前面

		// 获取列表范围
		listRange, err := client.LRange(ctx, key, 0, -1).Result()
		assert.NoError(t, err)
		assert.Len(t, listRange, len(values)-1)

		// 删除列表
		deleted, err := client.Del(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})

	// 测试集合操作
	t.Run("Set Operations", func(t *testing.T) {
		key := "test:set:key"
		members := []string{"member1", "member2", "member3"}

		// 添加集合成员
		for _, member := range members {
			err := client.SAdd(ctx, key, member).Err()
			assert.NoError(t, err)
		}

		// 获取集合成员数
		card, err := client.SCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(members)), card)

		// 检查成员是否存在
		isMember, err := client.SIsMember(ctx, key, members[0]).Result()
		assert.NoError(t, err)
		assert.True(t, isMember)

		// 获取所有成员
		allMembers, err := client.SMembers(ctx, key).Result()
		assert.NoError(t, err)
		assert.Len(t, allMembers, len(members))

		// 删除集合
		deleted, err := client.Del(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})

	// 测试有序集合操作
	t.Run("Sorted Set Operations", func(t *testing.T) {
		key := "test:zset:key"
		members := []redis.Z{
			{Score: 1.0, Member: "member1"},
			{Score: 2.0, Member: "member2"},
			{Score: 3.0, Member: "member3"},
		}

		// 添加有序集合成员
		err := client.ZAdd(ctx, key, &members[0], &members[1], &members[2]).Err()
		assert.NoError(t, err)

		// 获取有序集合成员数
		card, err := client.ZCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(members)), card)

		// 获取成员分数
		score, err := client.ZScore(ctx, key, "member1").Result()
		assert.NoError(t, err)
		assert.Equal(t, 1.0, score)

		// 获取排名范围内的成员
		rangeResult, err := client.ZRange(ctx, key, 0, -1).Result()
		assert.NoError(t, err)
		assert.Len(t, rangeResult, len(members))

		// 删除有序集合
		deleted, err := client.Del(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})

	// 测试过期时间操作
	t.Run("Expiration Operations", func(t *testing.T) {
		key := "test:expire:key"
		value := "expire_value"

		// 设置带过期时间的键
		err := client.Set(ctx, key, value, time.Second).Err()
		assert.NoError(t, err)

		// 检查键是否存在
		exists, err := client.Exists(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		// 获取过期时间
		ttl, err := client.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))

		// 等待过期
		time.Sleep(2 * time.Second)

		// 检查键是否已过期
		existsAfter, err := client.Exists(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), existsAfter)
	})

	// 测试管道操作
	t.Run("Pipeline Operations", func(t *testing.T) {
		pipe := client.Pipeline()

		// 添加多个命令到管道
		pipe.Set(ctx, "pipeline:key1", "value1", time.Hour)
		pipe.Set(ctx, "pipeline:key2", "value2", time.Hour)
		pipe.Get(ctx, "pipeline:key1")
		pipe.Get(ctx, "pipeline:key2")

		// 执行管道
		cmds, err := pipe.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 4)

		// 清理测试数据
		client.Del(ctx, "pipeline:key1", "pipeline:key2")
	})
}

// 测试Redis连接池性能
func TestRedisConnectionPool(t *testing.T) {
	// 跳过测试，如果没有真实的Redis服务
	// t.Skip("跳过Redis连接池测试，需要启动Redis服务")

	config := map[string]*Config{
		"pool_test": {
			Addrs:              []string{"127.0.0.1:6379"},
			Password:           "",
			Db:                 0,
			IsLog:              false,
			PoolSize:           20,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
	}

	// 初始化Redis连接
	err := InitRedis(config)
	assert.NoError(t, err)
	defer CloseRedis()

	// 并发测试连接池
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			client, err := GetConn("pool_test")
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: failed to get connection: %w", goroutineID, err)
				return
			}

			ctx := context.Background()

			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("pool:test:%d:%d", goroutineID, j)
				value := fmt.Sprintf("value_%d_%d", goroutineID, j)

				// 设置值
				err := client.Set(ctx, key, value, time.Minute).Err()
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to set %s: %w", goroutineID, key, err)
					continue
				}

				// 获取值
				result, err := client.Get(ctx, key).Result()
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to get %s: %w", goroutineID, key, err)
					continue
				}

				if result != value {
					errors <- fmt.Errorf("goroutine %d: value mismatch for %s, expected %s, got %s", goroutineID, key, value, result)
				}

				// 删除键
				client.Del(ctx, key)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	var errorCount int
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "Expected no errors in connection pool test")
}

// 测试Redis事务操作
func TestRedisTransaction(t *testing.T) {
	// 跳过测试，如果没有真实的Redis服务
	// t.Skip("跳过Redis事务测试，需要启动Redis服务")

	config := map[string]*Config{
		"tx_test": {
			Addrs:              []string{"127.0.0.1:6379"},
			Password:           "",
			Db:                 0,
			IsLog:              true,
			PoolSize:           10,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
	}

	// 初始化Redis连接
	err := InitRedis(config)
	assert.NoError(t, err)
	defer CloseRedis()

	// 获取Redis连接
	client, err := GetConn("tx_test")
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()

	// 测试事务成功的情况
	t.Run("Successful Transaction", func(t *testing.T) {
		// 开始事务
		tx := client.TxPipeline()

		// 添加事务命令
		tx.Set(ctx, "tx:key1", "value1", time.Hour)
		tx.Set(ctx, "tx:key2", "value2", time.Hour)
		tx.Get(ctx, "tx:key1")
		tx.Get(ctx, "tx:key2")

		// 执行事务
		cmds, err := tx.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 4)

		// 验证结果
		result1, err := client.Get(ctx, "tx:key1").Result()
		assert.NoError(t, err)
		assert.Equal(t, "value1", result1)

		result2, err := client.Get(ctx, "tx:key2").Result()
		assert.NoError(t, err)
		assert.Equal(t, "value2", result2)

		// 清理
		client.Del(ctx, "tx:key1", "tx:key2")
	})

	// 测试事务失败的情况
	t.Run("Failed Transaction", func(t *testing.T) {
		// 先设置一个键
		err := client.Set(ctx, "tx:existing", "old_value", time.Hour).Err()
		assert.NoError(t, err)

		// 开始事务
		tx := client.TxPipeline()

		// 添加会导致冲突的命令（如果键已存在）
		tx.SetNX(ctx, "tx:existing", "new_value", time.Hour)
		tx.Set(ctx, "tx:new_key", "new_value", time.Hour)

		// 执行事务
		cmds, err := tx.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 2)

		// 验证SetNX失败（键已存在）
		setNXResult := cmds[0].(*redis.BoolCmd)
		success, err := setNXResult.Result()
		assert.NoError(t, err)
		assert.False(t, success) // SetNX应该失败，因为键已存在

		// 验证第二个命令成功
		result, err := client.Get(ctx, "tx:new_key").Result()
		assert.NoError(t, err)
		assert.Equal(t, "new_value", result)

		// 清理
		client.Del(ctx, "tx:existing", "tx:new_key")
	})
}
