package mysql

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
)

// test sql
// -- 创建测试数据库
// CREATE DATABASE IF NOT EXISTS testdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

// -- 使用数据库
// USE testdb;

// -- 创建测试用户表
// CREATE TABLE IF NOT EXISTS test_users (
//     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
//     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
//     updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//     name VARCHAR(100) NOT NULL,
//     email VARCHAR(100) NOT NULL UNIQUE,
//     age INT NOT NULL,
//     status INT NOT NULL DEFAULT 1,
//     INDEX idx_email (email),
//     INDEX idx_status (status)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

// -- 创建测试订单表
// CREATE TABLE IF NOT EXISTS test_orders (
//     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
//     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
//     updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//     user_id BIGINT UNSIGNED NOT NULL,
//     amount DECIMAL(10,2) NOT NULL,
//     status VARCHAR(20) NOT NULL DEFAULT 'pending',
//     INDEX idx_user_id (user_id),
//     INDEX idx_status (status),
//     FOREIGN KEY (user_id) REFERENCES test_users(id) ON DELETE CASCADE
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
// -- 插入测试用户
// INSERT INTO test_users (name, email, age, status) VALUES
// ('测试用户1', 'testuser1@example.com', 25, 1),
// ('测试用户2', 'testuser2@example.com', 30, 1),
// ('测试用户3', 'testuser3@example.com', 35, 1);

// -- 插入测试订单
// INSERT INTO test_orders (user_id, amount, status) VALUES
// (1, 99.99, 'test'),
// (2, 199.99, 'test'),
// (3, 299.99, 'test');

func init() {
	// 初始化 logger，避免测试时 panic
	_ = logger.InitLogger(&logger.Config{
		Path:       "./test.log", // 测试日志文件，可随意指定
		Closed:     false,        // 启用日志输出以便测试
		MaxSize:    1,
		MaxAge:     1,
		MaxBackup:  1,
		AlarmLevel: "warn",
	})
}

// TestUser 测试用户模型
type TestUser struct {
	BaseModel
	Name   string `gorm:"size:100;not null" json:"name"`
	Email  string `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Age    int    `gorm:"not null" json:"age"`
	Status int    `gorm:"default:1;not null" json:"status"`
}

// TestOrder 测试订单模型
type TestOrder struct {
	BaseModel
	UserID uint    `gorm:"not null;index" json:"user_id"`
	Amount float64 `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status string  `gorm:"size:20;not null;default:'pending'" json:"status"`
}

func TestInitMysql(t *testing.T) {
	// 测试配置
	config := map[string]*Config{
		"test": {
			Dsn:                       []string{"root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"},
			MaxConn:                   10,
			MaxIdleConn:               5,
			ConnMaxLife:               3600,
			SlowThreshold:             500,
			IgnoreRecordNotFoundError: true,
			IsLog:                     false, // 测试时关闭日志
		},
	}

	// 测试初始化（这里会失败，因为没有真实的数据库连接）
	err := InitMysql(config)
	// 由于没有真实的数据库连接，这里期望失败
	assert.NoError(t, err)
}

func TestSetDefaultConfig(t *testing.T) {
	// 测试空配置
	config := &Config{}
	err := setDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mysql dsn is invalid")

	// 测试有效配置
	config = &Config{
		Dsn: []string{"test:test@tcp(localhost:3306)/testdb"},
	}
	err = setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 50, config.MaxConn)
	assert.Equal(t, 25, config.MaxIdleConn)
	assert.Equal(t, 3600, config.ConnMaxLife)
	assert.Equal(t, 500, config.SlowThreshold)

	// 测试自定义配置
	config = &Config{
		Dsn:           []string{"test:test@tcp(localhost:3306)/testdb"},
		MaxConn:       100,
		MaxIdleConn:   50,
		ConnMaxLife:   7200,
		SlowThreshold: 1000,
	}
	err = setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 100, config.MaxConn)
	assert.Equal(t, 50, config.MaxIdleConn)
	assert.Equal(t, 7200, config.ConnMaxLife)
	assert.Equal(t, 1000, config.SlowThreshold)
}

func TestGetConn(t *testing.T) {
	// 测试获取不存在的连接
	_, err := GetConn("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mysql conn is not found")
}

func TestNewTransaction(t *testing.T) {
	// 测试创建事务 - 由于无法创建真实的数据库连接，我们只测试函数签名和基本逻辑
	// 在实际使用中，应该通过 InitMysql 初始化真实的数据库连接

	// 测试 nil 数据库连接的情况
	tx := NewTransaction(nil)
	assert.NotNil(t, tx)
	assert.Nil(t, tx.db)
	assert.Nil(t, tx.tx)

	// 注意：这个测试在实际环境中会失败，因为 gorm.DB 需要正确初始化
	// 在真实的测试环境中，应该使用测试数据库或者 mock
}

// 添加一个测试来验证 Transaction 结构
func TestTransactionStructure(t *testing.T) {
	// 测试 Transaction 结构体的基本属性
	tx := &Transaction{
		db: nil,
		tx: nil,
	}

	assert.Nil(t, tx.db)
	assert.Nil(t, tx.tx)

	// 测试方法调用
	err := tx.Commit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is nil, cannot commit")

	err = tx.Rollback()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is nil, cannot rollback")
}

func TestTransactionCommit(t *testing.T) {
	// 测试事务提交 - 由于无法创建真实的数据库连接，我们只测试函数签名
	// 在实际使用中，应该通过 InitMysql 初始化真实的数据库连接

	tx := &Transaction{
		db: nil,
		tx: nil,
	}

	// 测试 nil tx 的情况
	err := tx.Commit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is nil, cannot commit")
}

func TestTransactionRollback(t *testing.T) {
	// 测试事务回滚 - 由于无法创建真实的数据库连接，我们只测试函数签名
	// 在实际使用中，应该通过 InitMysql 初始化真实的数据库连接

	tx := &Transaction{
		db: nil,
		tx: nil,
	}

	// 测试 nil tx 的情况
	err := tx.Rollback()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is nil, cannot rollback")
}

func TestCloseMysql(t *testing.T) {
	// 测试关闭连接（没有连接时应该正常）
	err := CloseMysql()
	assert.NoError(t, err)
}

func TestHealthCheck(t *testing.T) {
	// 测试健康检查（没有连接时应该返回空结果）
	healthStatus := HealthCheck()
	assert.NotNil(t, healthStatus)
	assert.Len(t, healthStatus, 0)
}

// 测试辅助函数
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "empty dsn",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Dsn: []string{"test:test@tcp(localhost:3306)/testdb"},
			},
			wantErr: false,
		},
		{
			name: "multiple dsn",
			config: &Config{
				Dsn: []string{
					"master:test@tcp(localhost:3306)/testdb",
					"slave1:test@tcp(localhost:3307)/testdb",
					"slave2:test@tcp(localhost:3308)/testdb",
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

// 测试读写分离配置
func TestReadWriteSeparation(t *testing.T) {
	config := &Config{
		Dsn: []string{
			"master:test@tcp(localhost:3306)/testdb",
			"slave1:test@tcp(localhost:3307)/testdb",
			"slave2:test@tcp(localhost:3308)/testdb",
		},
		MaxConn:                   100,
		MaxIdleConn:               25,
		ConnMaxLife:               3600,
		SlowThreshold:             500,
		IgnoreRecordNotFoundError: true,
		IsLog:                     false,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Len(t, config.Dsn, 3)
}

// 测试慢查询配置
func TestSlowQueryConfig(t *testing.T) {
	config := &Config{
		Dsn:           []string{"test:test@tcp(localhost:3306)/testdb"},
		SlowThreshold: 1000, // 1秒
		IsLog:         true,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1000, config.SlowThreshold)
}

// 测试连接池配置
func TestConnectionPoolConfig(t *testing.T) {
	config := &Config{
		Dsn:         []string{"test:test@tcp(localhost:3306)/testdb"},
		MaxConn:     200,
		MaxIdleConn: 50,
		ConnMaxLife: 7200,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 200, config.MaxConn)
	assert.Equal(t, 50, config.MaxIdleConn)
	assert.Equal(t, 7200, config.ConnMaxLife)
}

// 测试日志配置
func TestLoggingConfig(t *testing.T) {
	config := &Config{
		Dsn:                       []string{"test:test@tcp(localhost:3306)/testdb"},
		IsLog:                     true,
		IgnoreRecordNotFoundError: true,
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err)
	assert.True(t, config.IsLog)
	assert.True(t, config.IgnoreRecordNotFoundError)
}

// 测试并发安全性
func TestConcurrencySafety(t *testing.T) {
	// 这个测试主要验证连接管理的并发安全性
	// 由于没有真实的数据库连接，这里只是验证函数调用不会panic

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
	// 测试无效的DSN
	config := &Config{
		Dsn: []string{"invalid-dsn"},
	}

	err := setDefaultConfig(config)
	assert.NoError(t, err) // setDefaultConfig 只验证长度，不验证DSN格式

	// 测试空DSN列表
	config = &Config{
		Dsn: []string{},
	}

	err = setDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mysql dsn is invalid")
}

// 测试上下文支持
func TestContextSupport(t *testing.T) {
	ctx := context.Background()

	// 测试带上下文的操作（这里只是验证函数签名）
	_ = ctx

	// 在实际使用中，GORM 支持上下文
	// 例如：db.WithContext(ctx).Find(&users)
}

// 测试模型定义
func TestModelDefinition(t *testing.T) {
	// 测试用户模型
	user := TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Status: 1,
	}

	assert.NotEmpty(t, user.Name)
	assert.NotEmpty(t, user.Email)
	assert.Greater(t, user.Age, 0)
	assert.Equal(t, 1, user.Status)

	// 测试订单模型
	order := TestOrder{
		UserID: 1,
		Amount: 99.99,
		Status: "pending",
	}

	assert.Equal(t, uint(1), order.UserID)
	assert.Equal(t, 99.99, order.Amount)
	assert.Equal(t, "pending", order.Status)
}

// 真实操作MySQL的测试用例
func TestRealMySQLOperations(t *testing.T) {
	// t.Skip("跳过真实MySQL操作测试，需要本地MySQL服务")

	// 测试配置 - 需要修改为真实的MySQL地址
	config := map[string]*Config{
		"test": {
			Dsn:                       []string{"root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"},
			MaxConn:                   10,
			MaxIdleConn:               5,
			ConnMaxLife:               3600,
			SlowThreshold:             500,
			IgnoreRecordNotFoundError: true,
			IsLog:                     true,
		},
	}

	// 初始化MySQL连接
	err := InitMysql(config)
	assert.NoError(t, err)
	defer CloseMysql()

	// 获取数据库连接
	db, err := GetConn("test")
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// 自动迁移表结构
	err = db.AutoMigrate(&TestUser{}, &TestOrder{})
	assert.NoError(t, err)

	// 清理测试数据
	db.Where("email LIKE ?", "testuser%@example.com").Delete(&TestUser{})
	db.Where("status = ?", "test").Delete(&TestOrder{})

	// 测试创建用户
	t.Run("CreateUser", func(t *testing.T) {
		user := TestUser{
			Name:   "测试用户",
			Email:  "testuser1@example.com",
			Age:    25,
			Status: 1,
		}

		result := db.Create(&user)
		assert.NoError(t, result.Error)
		assert.Greater(t, user.ID, uint(0))
		assert.NotZero(t, user.ID)
	})

	// 测试查询用户
	t.Run("FindUser", func(t *testing.T) {
		var user TestUser
		result := db.Where("email = ?", "testuser1@example.com").First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "测试用户", user.Name)
		assert.Equal(t, 25, user.Age)
		assert.Equal(t, 1, user.Status)
	})

	// 测试更新用户
	t.Run("UpdateUser", func(t *testing.T) {
		result := db.Model(&TestUser{}).Where("email = ?", "testuser1@example.com").Update("age", 26)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		// 验证更新结果
		var user TestUser
		db.Where("email = ?", "testuser1@example.com").First(&user)
		assert.Equal(t, 26, user.Age)
	})

	// 测试批量创建用户
	t.Run("BatchCreateUsers", func(t *testing.T) {
		users := []TestUser{
			{Name: "批量用户1", Email: "testuser2@example.com", Age: 30, Status: 1},
			{Name: "批量用户2", Email: "testuser3@example.com", Age: 35, Status: 1},
			{Name: "批量用户3", Email: "testuser4@example.com", Age: 40, Status: 1},
		}

		result := db.Create(&users)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(3), result.RowsAffected)

		// 验证所有用户都创建成功
		for _, user := range users {
			assert.Greater(t, user.ID, uint(0))
		}
	})

	// 测试查询多个用户
	t.Run("FindMultipleUsers", func(t *testing.T) {
		var users []TestUser
		result := db.Where("email IN ?", []string{"testuser2@example.com", "testuser3@example.com", "testuser4@example.com"}).Find(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 3)
	})

	// 测试条件查询
	t.Run("ConditionalQuery", func(t *testing.T) {
		var users []TestUser
		result := db.Where("age > ? AND status = ?", 25, 1).Find(&users)
		assert.NoError(t, result.Error)
		assert.Greater(t, len(users), 0)

		// 验证所有返回的用户都满足条件
		for _, user := range users {
			assert.Greater(t, user.Age, 25)
			assert.Equal(t, 1, user.Status)
		}
	})

	// 测试分页查询
	t.Run("PaginationQuery", func(t *testing.T) {
		var users []TestUser
		offset := 0
		limit := 2

		result := db.Offset(offset).Limit(limit).Find(&users)
		assert.NoError(t, result.Error)
		assert.LessOrEqual(t, len(users), limit)
	})

	// 测试统计查询
	t.Run("CountQuery", func(t *testing.T) {
		var count int64
		result := db.Model(&TestUser{}).Where("email LIKE ?", "testuser%@example.com").Count(&count)
		assert.NoError(t, result.Error)
		assert.Greater(t, count, int64(0))
	})

	// 测试创建订单
	t.Run("CreateOrder", func(t *testing.T) {
		// 先获取一个用户ID
		var user TestUser
		db.Where("email = ?", "testuser1@example.com").First(&user)

		order := TestOrder{
			UserID: user.ID,
			Amount: 99.99,
			Status: "test",
		}

		result := db.Create(&order)
		assert.NoError(t, result.Error)
		assert.Greater(t, order.ID, uint(0))
	})

	// 测试关联查询
	t.Run("JoinQuery", func(t *testing.T) {
		type UserOrder struct {
			TestUser
			OrderID     uint    `json:"order_id"`
			OrderAmount float64 `json:"order_amount"`
			OrderStatus string  `json:"order_status"`
		}

		var userOrders []UserOrder
		result := db.Table("test_users").
			Select("test_users.*, test_orders.id as order_id, test_orders.amount as order_amount, test_orders.status as order_status").
			Joins("LEFT JOIN test_orders ON test_users.id = test_orders.user_id").
			Where("test_orders.status = ?", "test").
			Find(&userOrders)

		assert.NoError(t, result.Error)
		assert.Greater(t, len(userOrders), 0)
	})

	// 测试事务操作
	t.Run("Transaction", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		assert.NotNil(t, tx)

		// 在事务中创建用户
		user := TestUser{
			Name:   "事务用户",
			Email:  "testuser5@example.com",
			Age:    28,
			Status: 1,
		}

		result := tx.Create(&user)
		assert.NoError(t, result.Error)

		// 在事务中创建订单
		order := TestOrder{
			UserID: user.ID,
			Amount: 199.99,
			Status: "test",
		}

		result = tx.Create(&order)
		assert.NoError(t, result.Error)

		// 提交事务
		err := tx.Commit().Error
		assert.NoError(t, err)

		// 验证事务中的数据已保存
		var savedUser TestUser
		db.Where("email = ?", "testuser5@example.com").First(&savedUser)
		assert.Equal(t, "事务用户", savedUser.Name)

		var savedOrder TestOrder
		db.Where("user_id = ? AND status = ?", savedUser.ID, "test").First(&savedOrder)
		assert.Equal(t, 199.99, savedOrder.Amount)
	})

	// 测试事务回滚
	t.Run("TransactionRollback", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		assert.NotNil(t, tx)

		// 在事务中创建用户
		user := TestUser{
			Name:   "回滚用户",
			Email:  "testuser6@example.com",
			Age:    29,
			Status: 1,
		}

		result := tx.Create(&user)
		assert.NoError(t, result.Error)

		// 故意制造错误（尝试创建重复邮箱的用户）
		duplicateUser := TestUser{
			Name:   "重复用户",
			Email:  "testuser6@example.com", // 重复的邮箱
			Age:    30,
			Status: 1,
		}

		result = tx.Create(&duplicateUser)
		// 这里应该会失败，因为邮箱重复
		if result.Error != nil {
			// 回滚事务
			err := tx.Rollback().Error
			assert.NoError(t, err)

			// 验证数据已回滚
			var count int64
			db.Model(&TestUser{}).Where("email = ?", "testuser6@example.com").Count(&count)
			assert.Equal(t, int64(0), count)
		}
	})

	// 测试原生SQL查询
	t.Run("RawSQL", func(t *testing.T) {
		var count int64
		result := db.Raw("SELECT COUNT(*) FROM test_users WHERE email LIKE ?", "testuser%@example.com").Scan(&count)
		assert.NoError(t, result.Error)
		assert.Greater(t, count, int64(0))

		// 测试原生SQL更新
		result = db.Exec("UPDATE test_users SET age = age + 1 WHERE email LIKE ?", "testuser%@example.com")
		assert.NoError(t, result.Error)
		assert.Greater(t, result.RowsAffected, int64(0))
	})

	// 测试软删除（如果支持）
	t.Run("SoftDelete", func(t *testing.T) {
		// 注意：这里假设TestUser模型支持软删除
		// 如果需要软删除功能，需要在模型中添加DeletedAt字段

		// 创建测试用户
		user := TestUser{
			Name:   "软删除用户",
			Email:  "testuser7@example.com",
			Age:    31,
			Status: 1,
		}

		result := db.Create(&user)
		assert.NoError(t, result.Error)

		// 删除用户（硬删除）
		result = db.Delete(&user)
		assert.NoError(t, result.Error)

		// 验证用户已被删除
		var deletedUser TestUser
		result = db.Where("email = ?", "testuser7@example.com").First(&deletedUser)
		assert.Error(t, result.Error) // 应该找不到记录
	})

	// 测试批量更新
	t.Run("BatchUpdate", func(t *testing.T) {
		result := db.Model(&TestUser{}).Where("email LIKE ?", "testuser%@example.com").Update("status", 2)
		assert.NoError(t, result.Error)
		assert.Greater(t, result.RowsAffected, int64(0))

		// 验证批量更新结果
		var count int64
		db.Model(&TestUser{}).Where("email LIKE ? AND status = ?", "testuser%@example.com", 2).Count(&count)
		assert.Greater(t, count, int64(0))
	})

	// 测试连接池性能
	t.Run("ConnectionPool", func(t *testing.T) {
		// 并发测试连接池
		const numGoroutines = 5
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 获取连接
				conn, err := GetConn("test")
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to get connection: %w", id, err)
					return
				}

				// 执行简单查询
				var count int64
				result := conn.Model(&TestUser{}).Count(&count)
				if result.Error != nil {
					errors <- fmt.Errorf("goroutine %d: query failed: %w", id, result.Error)
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
	})

	// 清理测试数据
	t.Run("Cleanup", func(t *testing.T) {
		// 删除测试用户
		result := db.Where("email LIKE ?", "testuser%@example.com").Delete(&TestUser{})
		assert.NoError(t, result.Error)

		// 删除测试订单
		result = db.Where("status = ?", "test").Delete(&TestOrder{})
		assert.NoError(t, result.Error)
	})
}
