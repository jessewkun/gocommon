package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// User 用户模型示例
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Status   int    `json:"status"`
	CreateAt int64  `json:"create_at"`
}

// Order 订单模型示例
type Order struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status"`
	CreateAt int64   `json:"create_at"`
}

// ExampleRedisUsage Redis 使用示例
func ExampleRedisUsage() {
	// 1. 初始化 Redis 配置
	// 在真实应用中，这些配置通常来自配置文件，并通过config.Init()加载到Cfgs中
	// 这里为了演示，我们手动设置
	Cfgs = map[string]*Config{
		"default": {
			Addrs:              []string{"localhost:6379"},
			Password:           "",
			Db:                 0,
			IsLog:              true,
			PoolSize:           100,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       10,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
		"cluster": {
			Addrs:              []string{"localhost:7000", "localhost:7001", "localhost:7002"},
			Password:           "",
			Db:                 0,
			IsLog:              true,
			PoolSize:           50,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5,
			SlowThreshold:      100,
		},
	}

	// 2. 初始化 Redis 连接
	if err := Init(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	// 在示例结束时，确保关闭连接
	defer Close()

	// 3. 获取 Redis 连接
	client, err := GetConn("default")
	if err != nil {
		log.Fatalf("Failed to get Redis connection: %v", err)
	}

	ctx := context.Background()

	// 4. 基本字符串操作
	// 设置字符串
	if err := client.Set(ctx, "key1", "value1", time.Hour).Err(); err != nil {
		log.Printf("Failed to set key1: %v", err)
	} else {
		fmt.Println("Set key1 successfully")
	}

	// 获取字符串
	val, err := client.Get(ctx, "key1").Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Println("key1 does not exist")
		} else {
			log.Printf("Failed to get key1: %v", err)
		}
	} else {
		fmt.Printf("key1 value: %s\n", val)
	}

	// 5. 哈希表操作
	user := User{
		ID:       "user:1",
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      25,
		Status:   1,
		CreateAt: time.Now().Unix(),
	}

	// 设置哈希表
	if err := client.HSet(ctx, "user:1", map[string]interface{}{
		"name":      user.Name,
		"email":     user.Email,
		"age":       user.Age,
		"status":    user.Status,
		"create_at": user.CreateAt,
	}).Err(); err != nil {
		log.Printf("Failed to set user hash: %v", err)
	} else {
		fmt.Println("Set user hash successfully")
	}

	// 获取哈希表字段
	name, err := client.HGet(ctx, "user:1", "name").Result()
	if err != nil {
		log.Printf("Failed to get user name: %v", err)
	} else {
		fmt.Printf("User name: %s\n", name)
	}

	// 获取整个哈希表
	userData, err := client.HGetAll(ctx, "user:1").Result()
	if err != nil {
		log.Printf("Failed to get user data: %v", err)
	} else {
		fmt.Printf("User data: %+v\n", userData)
	}

	// 6. 列表操作
	// 从左侧推入元素
	if err := client.LPush(ctx, "list1", "item1", "item2", "item3").Err(); err != nil {
		log.Printf("Failed to push to list: %v", err)
	} else {
		fmt.Println("Pushed items to list successfully")
	}

	// 从右侧弹出元素
	item, err := client.RPop(ctx, "list1").Result()
	if err != nil {
		log.Printf("Failed to pop from list: %v", err)
	} else {
		fmt.Printf("Popped item: %s\n", item)
	}

	// 获取列表长度
	length, err := client.LLen(ctx, "list1").Result()
	if err != nil {
		log.Printf("Failed to get list length: %v", err)
	} else {
		fmt.Printf("List length: %d\n", length)
	}

	// 7. 集合操作
	// 添加集合元素
	if err := client.SAdd(ctx, "set1", "member1", "member2", "member3").Err(); err != nil {
		log.Printf("Failed to add set members: %v", err)
	} else {
		fmt.Println("Added set members successfully")
	}

	// 检查成员是否存在
	exists, err := client.SIsMember(ctx, "set1", "member1").Result()
	if err != nil {
		log.Printf("Failed to check set member: %v", err)
	} else {
		fmt.Printf("member1 exists in set1: %t\n", exists)
	}

	// 获取集合所有成员
	members, err := client.SMembers(ctx, "set1").Result()
	if err != nil {
		log.Printf("Failed to get set members: %v", err)
	} else {
		fmt.Printf("Set members: %+v\n", members)
	}

	// 8. 有序集合操作
	// 添加有序集合元素
	if err := client.ZAdd(ctx, "zset1", &redis.Z{Score: 1.0, Member: "member1"},
		&redis.Z{Score: 2.0, Member: "member2"},
		&redis.Z{Score: 3.0, Member: "member3"}).Err(); err != nil {
		log.Printf("Failed to add zset members: %v", err)
	} else {
		fmt.Println("Added zset members successfully")
	}

	// 获取有序集合成员（按分数升序）
	zmembers, err := client.ZRange(ctx, "zset1", 0, -1).Result()
	if err != nil {
		log.Printf("Failed to get zset members: %v", err)
	} else {
		fmt.Printf("ZSet members: %+v\n", zmembers)
	}

	// 9. 过期时间操作
	// 设置过期时间
	if err := client.Expire(ctx, "key1", time.Minute).Err(); err != nil {
		log.Printf("Failed to set expire: %v", err)
	} else {
		fmt.Println("Set expire successfully")
	}

	// 获取剩余过期时间
	ttl, err := client.TTL(ctx, "key1").Result()
	if err != nil {
		log.Printf("Failed to get TTL: %v", err)
	} else {
		fmt.Printf("TTL: %v\n", ttl)
	}

	// 10. 管道操作（批量执行）
	pipe := client.Pipeline()
	pipe.Set(ctx, "pipe_key1", "value1", time.Hour)
	pipe.Set(ctx, "pipe_key2", "value2", time.Hour)
	pipe.Get(ctx, "pipe_key1")
	pipe.Get(ctx, "pipe_key2")

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("Failed to execute pipeline: %v", err)
	} else {
		fmt.Printf("Pipeline executed successfully, %d commands\n", len(cmds))
	}

	// 11. 事务操作
	txf := func(tx *redis.Tx) error {
		// 获取当前值
		val, err := tx.Get(ctx, "counter").Result()
		if err != nil && err != redis.Nil {
			return err
		}

		// 模拟一些处理时间
		time.Sleep(time.Millisecond * 100)

		// 在事务中执行操作
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if val == "" {
				pipe.Set(ctx, "counter", 1, time.Hour)
			} else {
				pipe.Incr(ctx, "counter")
			}
			return nil
		})
		return err
	}

	// 执行事务
	for i := 0; i < 3; i++ {
		err := client.Watch(ctx, txf, "counter")
		if err == nil {
			break
		}
		if err == redis.TxFailedErr {
			continue
		}
		log.Printf("Transaction failed: %v", err)
		break
	}

	// 12. 发布订阅
	// 创建订阅
	pubsub := client.Subscribe(ctx, "channel1")
	defer pubsub.Close()

	// 发布消息
	if err := client.Publish(ctx, "channel1", "Hello Redis!").Err(); err != nil {
		log.Printf("Failed to publish message: %v", err)
	} else {
		fmt.Println("Published message successfully")
	}

	// 接收消息（非阻塞）
	msg, err := pubsub.ReceiveTimeout(ctx, time.Second)
	if err != nil {
		log.Printf("Failed to receive message: %v", err)
	} else {
		fmt.Printf("Received message: %+v\n", msg)
	}

	// 13. 键操作
	// 检查键是否存在
	existsCount, err := client.Exists(ctx, "key1", "key2").Result()
	if err != nil {
		log.Printf("Failed to check key existence: %v", err)
	} else {
		fmt.Printf("Number of existing keys: %d\n", existsCount)
	}

	// 删除键
	deleted, err := client.Del(ctx, "key1").Result()
	if err != nil {
		log.Printf("Failed to delete key: %v", err)
	} else {
		fmt.Printf("Deleted %d keys\n", deleted)
	}

	// 14. 数据库操作
	// 清空当前数据库
	if err := client.FlushDB(ctx).Err(); err != nil {
		log.Printf("Failed to flush database: %v", err)
	} else {
		fmt.Println("Database flushed successfully")
	}

	// 15. 健康检查
	healthStatus := HealthCheck()
	for dbName, status := range healthStatus {
		fmt.Printf("Redis %s health status: %+v\n", dbName, status)
	}

	// 16. 关闭连接
	// 注意：Redis 模块可能没有 Close 函数，这里只是示例
	fmt.Println("Redis example completed successfully")
}
