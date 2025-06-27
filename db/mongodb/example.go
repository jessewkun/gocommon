package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// User 用户模型示例
type User struct {
	ID      string    `bson:"_id,omitempty"`
	Name    string    `bson:"name"`
	Email   string    `bson:"email"`
	Age     int       `bson:"age"`
	Created time.Time `bson:"created"`
	Updated time.Time `bson:"updated"`
}

// ExampleMongoDBUsage MongoDB 使用示例
func ExampleMongoDBUsage() {
	// 1. 设置 MongoDB 配置
	// 在实际应用中，这些配置通常通过配置文件加载
	Cfgs = map[string]*Config{
		"default": {
			Uris:                   []string{"mongodb://localhost:27017"},
			MaxPoolSize:            100,
			MinPoolSize:            5,
			MaxConnIdleTime:        300,
			ConnectTimeout:         10,
			ServerSelectionTimeout: 5,
			SocketTimeout:          30,
			ReadPreference:         "primary",
			WriteConcern:           "majority",
		},
		"replica": {
			Uris:                   []string{"mongodb://localhost:27017,localhost:27018,localhost:27019"},
			MaxPoolSize:            50,
			MinPoolSize:            3,
			MaxConnIdleTime:        300,
			ConnectTimeout:         10,
			ServerSelectionTimeout: 5,
			SocketTimeout:          30,
			ReadPreference:         "secondaryPreferred",
			WriteConcern:           "majority",
		},
	}

	// 2. 初始化 MongoDB 连接
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// 3. 获取集合
	collection, err := GetCollection("default", "testdb", "users")
	if err != nil {
		log.Fatalf("Failed to get collection: %v", err)
	}

	// 4. 插入文档
	user := User{
		Name:    "张三",
		Email:   "zhangsan@example.com",
		Age:     25,
		Created: time.Now(),
		Updated: time.Now(),
	}

	insertResult, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Printf("Failed to insert document: %v", err)
	} else {
		fmt.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)
	}

	// 5. 查询文档
	var foundUser User
	err = collection.FindOne(context.Background(), bson.M{"name": "张三"}).Decode(&foundUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found for name '张三'")
		} else {
			log.Printf("Failed to find document: %v", err)
		}
	} else {
		fmt.Printf("Found user with name '张三': %+v\n", foundUser)
	}

	// 6. 更新文档
	update := bson.M{
		"$set": bson.M{
			"age":     26,
			"updated": time.Now(),
		},
	}

	updateResult, err := collection.UpdateOne(
		context.Background(),
		bson.M{"name": "张三"},
		update,
	)
	if err != nil {
		log.Printf("Failed to update document: %v", err)
	} else {
		fmt.Printf("Updated %v document(s) for name '张三'\n", updateResult.ModifiedCount)
	}

	// 7. 删除文档
	deleteResult, err := collection.DeleteOne(context.Background(), bson.M{"name": "张三"})
	if err != nil {
		log.Printf("Failed to delete document: %v", err)
	} else {
		fmt.Printf("Deleted %v document(s) for name '张三'\n", deleteResult.DeletedCount)
	}

	// 8. 使用事务
	client, err := GetConn("default")
	if err != nil {
		log.Fatalf("Failed to get client for transaction: %v", err)
	}

	// 事务操作需要副本集或 mongos
	// 请确保您的 MongoDB 环境支持事务
	err = WithTransaction(client, func(sessCtx mongo.SessionContext) error {
		// 在事务中执行操作
		_, err := collection.InsertOne(sessCtx, User{
			Name:    "李四",
			Email:   "lisi@example.com",
			Age:     30,
			Created: time.Now(),
			Updated: time.Now(),
		})
		if err != nil {
			return err
		}

		_, err = collection.InsertOne(sessCtx, User{
			Name:    "王五",
			Email:   "wangwu@example.com",
			Age:     28,
			Created: time.Now(),
			Updated: time.Now(),
		})
		return err
	})

	if err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			// 检查是否是由于环境不支持事务导致的错误
			isTransactionError := false
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 20 { // 事务错误码
					isTransactionError = true
					break
				}
			}
			if isTransactionError {
				fmt.Println("Transaction failed: The current MongoDB deployment does not support transactions.")
			} else {
				log.Printf("Transaction failed with write exception: %v", err)
			}
		} else {
			log.Printf("Transaction failed: %v", err)
		}
	} else {
		fmt.Println("Transaction completed successfully")
	}

	// 9. 批量操作
	users := []interface{}{
		User{Name: "赵六", Email: "zhaoliu@example.com", Age: 22, Created: time.Now(), Updated: time.Now()},
		User{Name: "钱七", Email: "qianqi@example.com", Age: 35, Created: time.Now(), Updated: time.Now()},
		User{Name: "孙八", Email: "sunba@example.com", Age: 29, Created: time.Now(), Updated: time.Now()},
	}

	insertManyResult, err := collection.InsertMany(context.Background(), users)
	if err != nil {
		log.Printf("Failed to insert many documents: %v", err)
	} else {
		fmt.Printf("Inserted %v documents\n", len(insertManyResult.InsertedIDs))
	}

	// 10. 聚合查询
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"age": bson.M{"$gte": 25}}}},
		{{Key: "$group", Value: bson.M{
			"_id":    nil,
			"avgAge": bson.M{"$avg": "$age"},
			"count":  bson.M{"$sum": 1},
		}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Printf("Failed to execute aggregation: %v", err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Printf("Failed to decode aggregation results: %v", err)
	}

	fmt.Printf("Aggregation results: %+v\n", results)

	// 11. 健康检查
	healthStatus := HealthCheck()
	for dbName, status := range healthStatus {
		fmt.Printf("MongoDB '%s' health status: %+v\n", dbName, status)
	}

	// 12. 关闭连接
	if err := Close(); err != nil {
		log.Printf("Failed to close MongoDB connections: %v", err)
	} else {
		fmt.Println("All MongoDB connections closed successfully.")
	}
}
