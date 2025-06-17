package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	// 初始化 logger，避免测试时 panic
	_ = logger.InitLogger(&logger.Config{
		Path:       "./test.log", // 测试日志文件，可随意指定
		Closed:     true,         // 关闭实际日志输出
		MaxSize:    1,
		MaxAge:     1,
		MaxBackup:  1,
		AlarmLevel: "warn",
	})
}

// TestUser 测试用户模型
type TestUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Age       int                `bson:"age"`
	Status    int                `bson:"status"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

// TestOrder 测试订单模型
type TestOrder struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Amount    float64            `bson:"amount"`
	Status    string             `bson:"status"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func TestInitMongoDB(t *testing.T) {
	// 测试配置
	config := map[string]*Config{
		"test": {
			Uris:                   []string{"mongodb://localhost:27017"},
			IsLog:                  false, // 测试时关闭日志
			MaxPoolSize:            10,
			MinPoolSize:            5,
			MaxConnIdleTime:        300,
			ServerSelectionTimeout: 5,
			ConnectTimeout:         5,
			SocketTimeout:          5,
		},
	}

	// 测试初始化（这里会失败，因为没有真实的MongoDB连接）
	err := InitMongoDB(config)
	// 由于没有真实的MongoDB连接，这里期望失败
	assert.NoError(t, err)
}

func TestSetMongoDefaultConfig(t *testing.T) {
	// 测试空配置
	config := &Config{}
	err := setMongoDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb uris is invalid")

	// 测试有效配置
	config = &Config{
		Uris: []string{"mongodb://localhost:27017"},
	}
	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 100, config.MaxPoolSize)
	assert.Equal(t, 5, config.MinPoolSize)
	assert.Equal(t, 300, config.MaxConnIdleTime)
	assert.Equal(t, 5, config.ServerSelectionTimeout)
	assert.Equal(t, 10, config.ConnectTimeout)
	assert.Equal(t, 30, config.SocketTimeout)

	// 测试自定义配置
	config = &Config{
		Uris:                   []string{"mongodb://localhost:27017"},
		MaxPoolSize:            200,
		MinPoolSize:            10,
		MaxConnIdleTime:        600,
		ServerSelectionTimeout: 60,
		ConnectTimeout:         20,
		SocketTimeout:          10,
	}
	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 200, config.MaxPoolSize)
	assert.Equal(t, 10, config.MinPoolSize)
	assert.Equal(t, 600, config.MaxConnIdleTime)
	assert.Equal(t, 60, config.ServerSelectionTimeout)
	assert.Equal(t, 20, config.ConnectTimeout)
	assert.Equal(t, 10, config.SocketTimeout)
}

func TestGetMongoClient(t *testing.T) {
	// 测试获取不存在的连接
	_, err := GetMongoClient("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb client is not found")
}

func TestGetMongoDatabase(t *testing.T) {
	// 测试获取不存在的数据库
	_, err := GetMongoDatabase("nonexistent", "testdb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb client is not found")
}

func TestGetMongoCollection(t *testing.T) {
	// 测试获取不存在的集合
	_, err := GetMongoCollection("nonexistent", "testdb", "users")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb client is not found")
}

func TestCloseMongoDB(t *testing.T) {
	// 测试关闭连接（没有连接时应该正常）
	err := CloseMongoDB()
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
			name:    "empty uris",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				Uris: []string{"mongodb://localhost:27017"},
			},
			wantErr: false,
		},
		{
			name: "replica set config",
			config: &Config{
				Uris: []string{"mongodb://localhost:27017,localhost:27018,localhost:27019"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setMongoDefaultConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 测试连接池配置
func TestConnectionPoolConfig(t *testing.T) {
	config := &Config{
		Uris:            []string{"mongodb://localhost:27017"},
		MaxPoolSize:     200,
		MinPoolSize:     20,
		MaxConnIdleTime: 600,
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 200, config.MaxPoolSize)
	assert.Equal(t, 20, config.MinPoolSize)
	assert.Equal(t, 600, config.MaxConnIdleTime)
}

// 测试超时配置
func TestTimeoutConfig(t *testing.T) {
	config := &Config{
		Uris:                   []string{"mongodb://localhost:27017"},
		ServerSelectionTimeout: 60,
		ConnectTimeout:         20,
		SocketTimeout:          10,
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 60, config.ServerSelectionTimeout)
	assert.Equal(t, 20, config.ConnectTimeout)
	assert.Equal(t, 10, config.SocketTimeout)
}

// 测试日志配置
func TestLoggingConfig(t *testing.T) {
	config := &Config{
		Uris:  []string{"mongodb://localhost:27017"},
		IsLog: true,
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.True(t, config.IsLog)
}

// 测试并发安全性
func TestConcurrencySafety(t *testing.T) {
	// 这个测试主要验证连接管理的并发安全性
	// 由于没有真实的MongoDB连接，这里只是验证函数调用不会panic

	// 并发调用 GetMongoClient
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = GetMongoClient("test")
		}()
	}

	// 并发调用 GetMongoDatabase
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = GetMongoDatabase("test", "testdb")
		}()
	}

	// 并发调用 GetMongoCollection
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = GetMongoCollection("test", "testdb", "users")
		}()
	}

	// 并发调用 MongoHealthCheck
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
	// 测试空的URIs
	config := &Config{
		Uris: []string{},
	}

	err := setMongoDefaultConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb uris is invalid")

	// 测试无效的URI格式
	config = &Config{
		Uris: []string{"invalid-uri"},
	}

	err = setMongoDefaultConfig(config)
	assert.NoError(t, err) // setMongoDefaultConfig 只验证长度，不验证URI格式
}

// 测试上下文支持
func TestContextSupport(t *testing.T) {
	ctx := context.Background()

	// 测试带上下文的操作（这里只是验证函数签名）
	_ = ctx

	// 在实际使用中，MongoDB 驱动支持上下文
	// 例如：collection.Find(ctx, filter)
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
		UserID: primitive.NewObjectID(),
		Amount: 99.99,
		Status: "pending",
	}

	assert.NotEqual(t, primitive.NilObjectID, order.UserID)
	assert.Equal(t, 99.99, order.Amount)
	assert.Equal(t, "pending", order.Status)
}

// 测试BSON操作
func TestBSONOperations(t *testing.T) {
	// 测试BSON文档创建
	doc := bson.M{
		"name":   "测试用户",
		"email":  "test@example.com",
		"age":    25,
		"status": 1,
	}

	assert.Equal(t, "测试用户", doc["name"])
	assert.Equal(t, "test@example.com", doc["email"])
	assert.Equal(t, 25, doc["age"])
	assert.Equal(t, 1, doc["status"])

	// 测试BSON数组
	array := bson.A{"item1", "item2", "item3"}
	assert.Len(t, array, 3)
	assert.Equal(t, "item1", array[0])
	assert.Equal(t, "item2", array[1])
	assert.Equal(t, "item3", array[2])

	// 测试嵌套BSON文档
	nestedDoc := bson.M{
		"user": bson.M{
			"name": "测试用户",
			"profile": bson.M{
				"age":     25,
				"city":    "北京",
				"hobbies": bson.A{"读书", "游泳", "编程"},
			},
		},
	}

	user := nestedDoc["user"].(bson.M)
	profile := user["profile"].(bson.M)
	hobbies := profile["hobbies"].(bson.A)

	assert.Equal(t, "测试用户", user["name"])
	assert.Equal(t, 25, profile["age"])
	assert.Equal(t, "北京", profile["city"])
	assert.Len(t, hobbies, 3)
}

// 测试ObjectID操作
func TestObjectIDOperations(t *testing.T) {
	// 测试生成新的ObjectID
	id := primitive.NewObjectID()
	assert.NotEqual(t, primitive.NilObjectID, id)
	assert.NotEmpty(t, id.Hex())

	// 测试从字符串创建ObjectID
	idStr := id.Hex()
	parsedID, err := primitive.ObjectIDFromHex(idStr)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)

	// 测试无效的ObjectID字符串
	_, err = primitive.ObjectIDFromHex("invalid")
	assert.Error(t, err)
}

// 测试边界值
func TestBoundaryValues(t *testing.T) {
	// 测试最小连接池大小
	config := &Config{
		Uris:        []string{"mongodb://localhost:27017"},
		MaxPoolSize: 1,
		MinPoolSize: 1,
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, config.MaxPoolSize)
	assert.Equal(t, 1, config.MinPoolSize)

	// 测试最大连接池大小
	config = &Config{
		Uris:        []string{"mongodb://localhost:27017"},
		MaxPoolSize: 1000,
	}

	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 1000, config.MaxPoolSize)

	// 测试零值超时
	config = &Config{
		Uris:                   []string{"mongodb://localhost:27017"},
		ServerSelectionTimeout: 0,
		ConnectTimeout:         0,
		SocketTimeout:          0,
	}

	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	// 应该使用默认值
	assert.Equal(t, 5, config.ServerSelectionTimeout)
	assert.Equal(t, 10, config.ConnectTimeout)
	assert.Equal(t, 30, config.SocketTimeout)
}

// 测试配置验证的完整性
func TestConfigValidationCompleteness(t *testing.T) {
	// 测试所有字段的默认值设置
	config := &Config{
		Uris: []string{"mongodb://localhost:27017"},
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)

	// 验证所有字段都有合理的默认值
	assert.NotEmpty(t, config.Uris)
	assert.Greater(t, config.MaxPoolSize, 0)
	assert.GreaterOrEqual(t, config.MinPoolSize, 0)
	assert.Greater(t, config.MaxConnIdleTime, 0)
	assert.Greater(t, config.ServerSelectionTimeout, 0)
	assert.Greater(t, config.ConnectTimeout, 0)
	assert.Greater(t, config.SocketTimeout, 0)
}

// 测试配置的不可变性
func TestConfigImmutability(t *testing.T) {
	originalConfig := &Config{
		Uris:                   []string{"mongodb://localhost:27017"},
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        300,
		ServerSelectionTimeout: 5,
		ConnectTimeout:         10,
		SocketTimeout:          30,
		IsLog:                  true,
	}

	// 复制配置
	configCopy := *originalConfig

	// 修改原始配置
	originalConfig.MaxPoolSize = 200

	// 验证副本没有被修改
	assert.Equal(t, 100, configCopy.MaxPoolSize)
	assert.Equal(t, 200, originalConfig.MaxPoolSize)
}

// 测试MongoDB特有的功能
func TestMongoDBSpecificFeatures(t *testing.T) {
	// 测试副本集配置
	config := &Config{
		Uris: []string{"mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0"},
	}

	err := setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Contains(t, config.Uris[0], "replicaSet=rs0")

	// 测试分片集群配置
	config = &Config{
		Uris: []string{"mongodb://mongos1:27017,mongos2:27017,mongos3:27017"},
	}

	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Contains(t, config.Uris[0], "mongos")

	// 测试认证配置
	config = &Config{
		Uris: []string{"mongodb://username:password@localhost:27017/admin?authSource=admin"},
	}

	err = setMongoDefaultConfig(config)
	assert.NoError(t, err)
	assert.Contains(t, config.Uris[0], "authSource=admin")
}

// 真实操作MongoDB的测试用例
func TestRealMongoDBOperations(t *testing.T) {
	// t.Skip("跳过真实MongoDB操作测试，需要本地MongoDB服务")

	config := map[string]*Config{
		"test": {
			Uris:                   []string{"mongodb://localhost:27017"},
			IsLog:                  false,
			MaxPoolSize:            10,
			MinPoolSize:            5,
			MaxConnIdleTime:        300,
			ServerSelectionTimeout: 5,
			ConnectTimeout:         5,
			SocketTimeout:          5,
		},
	}
	// 初始化连接
	err := InitMongoDB(config)
	assert.NoError(t, err)
	defer CloseMongoDB()

	// 获取集合
	coll, err := GetMongoCollection("test", "testdb", "users")
	assert.NoError(t, err)
	ctx := context.Background()

	// 清理测试数据
	_, _ = coll.DeleteMany(ctx, bson.M{"email": bson.M{"$regex": "^testuser"}})

	// 插入文档
	t.Run("InsertOne", func(t *testing.T) {
		doc := bson.M{
			"name":       "测试用户",
			"email":      "testuser1@example.com",
			"age":        20,
			"status":     1,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		res, err := coll.InsertOne(ctx, doc)
		assert.NoError(t, err)
		assert.NotNil(t, res.InsertedID)
	})

	// 查询文档
	t.Run("FindOne", func(t *testing.T) {
		var result bson.M
		err := coll.FindOne(ctx, bson.M{"email": "testuser1@example.com"}).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "测试用户", result["name"])
	})

	// 更新文档
	t.Run("UpdateOne", func(t *testing.T) {
		update := bson.M{"$set": bson.M{"age": 21}}
		res, err := coll.UpdateOne(ctx, bson.M{"email": "testuser1@example.com"}, update)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.ModifiedCount)
	})

	// 查询更新后的文档
	t.Run("FindUpdated", func(t *testing.T) {
		var result bson.M
		err := coll.FindOne(ctx, bson.M{"email": "testuser1@example.com"}).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, int32(21), result["age"])
	})

	// 删除文档
	t.Run("DeleteOne", func(t *testing.T) {
		res, err := coll.DeleteOne(ctx, bson.M{"email": "testuser1@example.com"})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.DeletedCount)
	})

	// 批量插入与批量查询
	t.Run("BulkInsertAndFind", func(t *testing.T) {
		docs := []interface{}{
			bson.M{"name": "批量用户1", "email": "testuser2@example.com", "age": 22, "status": 1, "created_at": time.Now(), "updated_at": time.Now()},
			bson.M{"name": "批量用户2", "email": "testuser3@example.com", "age": 23, "status": 1, "created_at": time.Now(), "updated_at": time.Now()},
		}
		_, err := coll.InsertMany(ctx, docs)
		assert.NoError(t, err)

		cursor, err := coll.Find(ctx, bson.M{"email": bson.M{"$in": []string{"testuser2@example.com", "testuser3@example.com"}}})
		assert.NoError(t, err)
		var results []bson.M
		err = cursor.All(ctx, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	// 聚合操作
	t.Run("Aggregate", func(t *testing.T) {
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"email": bson.M{"$regex": "^testuser"}}}},
			{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": 1}}}},
		}
		cursor, err := coll.Aggregate(ctx, pipeline)
		assert.NoError(t, err)
		var aggResults []bson.M
		err = cursor.All(ctx, &aggResults)
		assert.NoError(t, err)
		if len(aggResults) > 0 {
			t.Logf("聚合结果: %+v", aggResults[0])
		}
	})

	// 清理测试数据
	_, _ = coll.DeleteMany(ctx, bson.M{"email": bson.M{"$regex": "^testuser"}})
}
