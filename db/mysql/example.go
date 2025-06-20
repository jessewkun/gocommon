package mysql

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// User 用户模型示例
type User struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:100;uniqueIndex;not null"`
	Age       int       `gorm:"not null"`
	Status    int       `gorm:"default:1;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// Order 订单模型示例
type Order struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null;index"`
	Amount    float64   `gorm:"type:decimal(10,2);not null"`
	Status    string    `gorm:"size:20;not null;default:'pending'"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// ExampleMysqlUsage MySQL 使用示例
func ExampleMysqlUsage() {
	// 1. 初始化 MySQL 配置
	// 在真实应用中，这些配置通常来自配置文件，并通过config.Init()加载到Cfgs中
	// 这里为了演示，我们手动设置
	Cfgs = map[string]*Config{
		"default": {
			Dsn:                       []string{"root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"},
			MaxConn:                   100,
			MaxIdleConn:               25,
			ConnMaxLife:               3600,
			SlowThreshold:             500,
			IgnoreRecordNotFoundError: true,
			IsLog:                     true,
		},
		"slave": {
			Dsn: []string{
				"root:password@tcp(master:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
				"root:password@tcp(slave1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
				"root:password@tcp(slave2:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
			},
			MaxConn:                   50,
			MaxIdleConn:               10,
			ConnMaxLife:               3600,
			SlowThreshold:             1000,
			IgnoreRecordNotFoundError: true,
			IsLog:                     true,
		},
	}

	// 2. 初始化 MySQL 连接
	if err := InitMysql(); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}
	// 在示例结束时，确保关闭连接
	defer CloseMysql()

	// 3. 获取数据库连接
	db, err := GetConn("default")
	if err != nil {
		log.Fatalf("Failed to get MySQL connection: %v", err)
	}

	// 4. 自动迁移表结构
	if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
		log.Printf("Failed to auto migrate: %v", err)
	}

	// 5. 插入数据
	user := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		fmt.Printf("Created user with ID: %d\n", user.ID)
	}

	// 6. 查询数据
	var foundUser User
	if err := db.Where("email = ?", "zhangsan@example.com").First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("User not found")
		} else {
			log.Printf("Failed to find user: %v", err)
		}
	} else {
		fmt.Printf("Found user: %+v\n", foundUser)
	}

	// 7. 更新数据
	if err := db.Model(&foundUser).Updates(map[string]interface{}{
		"age":        26,
		"updated_at": time.Now(),
	}).Error; err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		fmt.Printf("Updated user age to: %d\n", foundUser.Age)
	}

	// 8. 删除数据
	if err := db.Delete(&foundUser).Error; err != nil {
		log.Printf("Failed to delete user: %v", err)
	} else {
		fmt.Printf("Deleted user with ID: %d\n", foundUser.ID)
	}

	// 9. 批量操作
	users := []User{
		{Name: "李四", Email: "lisi@example.com", Age: 30, Status: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "王五", Email: "wangwu@example.com", Age: 28, Status: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 22, Status: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Printf("Failed to create users: %v", err)
	} else {
		fmt.Printf("Created %d users\n", len(users))
	}

	// 10. 复杂查询
	var userList []User
	if err := db.Where("age >= ? AND status = ?", 25, 1).Find(&userList).Error; err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		fmt.Printf("Found %d users with age >= 25\n", len(userList))
	}

	// 11. 关联查询示例
	// 创建订单
	order := Order{
		UserID:    foundUser.ID,
		Amount:    99.99,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&order).Error; err != nil {
		log.Printf("Failed to create order: %v", err)
	}

	// 查询用户及其订单
	var userWithOrders User
	if err := db.Preload("Orders").Where("id = ?", foundUser.ID).First(&userWithOrders).Error; err != nil {
		log.Printf("Failed to query user with orders: %v", err)
	}

	// 12. 事务处理
	tx := NewTransaction(db)
	if err := tx.tx.Begin().Error; err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// 在事务中执行操作
	newUser := User{
		Name:      "事务用户",
		Email:     "transaction@example.com",
		Age:       35,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to create user in transaction: %v", err)
		return
	}

	newOrder := Order{
		UserID:    newUser.ID,
		Amount:    199.99,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to create order in transaction: %v", err)
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}

	// 13. 原生 SQL 查询
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM users WHERE age >= ?", 25).Scan(&count).Error; err != nil {
		log.Printf("Failed to execute raw SQL: %v", err)
	} else {
		fmt.Printf("Total users with age >= 25: %d\n", count)
	}

	// 14. 分页查询
	var pageUsers []User
	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize

	if err := db.Offset(offset).Limit(pageSize).Find(&pageUsers).Error; err != nil {
		log.Printf("Failed to query users with pagination: %v", err)
	} else {
		fmt.Printf("Page %d users: %d\n", page, len(pageUsers))
	}

	// 15. 聚合查询
	type UserStats struct {
		AvgAge float64 `json:"avg_age"`
		MaxAge int     `json:"max_age"`
		MinAge int     `json:"min_age"`
		Count  int64   `json:"count"`
	}

	var stats UserStats
	if err := db.Model(&User{}).Select("AVG(age) as avg_age, MAX(age) as max_age, MIN(age) as min_age, COUNT(*) as count").Scan(&stats).Error; err != nil {
		log.Printf("Failed to get user stats: %v", err)
	} else {
		fmt.Printf("User stats: %+v\n", stats)
	}

	// 16. 健康检查
	healthStatus := HealthCheck()
	for dbName, status := range healthStatus {
		fmt.Printf("MySQL %s health status: %+v\n", dbName, status)
	}

	// 17. 关闭连接
	if err := CloseMysql(); err != nil {
		log.Printf("Failed to close MySQL connections: %v", err)
	} else {
		fmt.Println("MySQL connections closed successfully")
	}
}
