package mongodb

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	gocommonlog "github.com/jessewkun/gocommon/logger"
)

const (
	testDBInstance   = "default"
	testDatabaseName = "testdb"
	testCollection   = "testusers"
	mongoURI         = "mongodb://localhost:27017"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set up logger configuration for tests
	gocommonlog.Cfg.Path = "./test.log"
	gocommonlog.Cfg.Closed = false
	gocommonlog.Cfg.MaxSize = 100
	gocommonlog.Cfg.MaxAge = 30
	gocommonlog.Cfg.MaxBackup = 10
	gocommonlog.Cfg.AlarmLevel = "warn"

	// Initialize logger first
	if err := gocommonlog.InitLogger(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Set up configuration for the test MongoDB instance
	Cfgs = map[string]*Config{
		testDBInstance: {
			Uris:                   []string{mongoURI},
			MaxPoolSize:            10,
			MinPoolSize:            1,
			MaxConnIdleTime:        60,
			ConnectTimeout:         5,
			ServerSelectionTimeout: 3,
			SocketTimeout:          10,
			ReadPreference:         "primary",
			WriteConcern:           "majority",
			IsLog:                  false,
		},
	}

	// Initialize MongoDB
	if err := InitMongoDB(); err != nil {
		fmt.Printf("Failed to initialize MongoDB for tests: %v\n", err)
		// Check if the error indicates a connection problem
		if _, ok := err.(mongo.ServerError); ok || err.Error() == "server selection error: context deadline exceeded" {
			fmt.Println("Skipping MongoDB tests: Cannot connect to MongoDB instance at", mongoURI)
			os.Exit(0) // Exit gracefully, skipping tests
		}
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	// Teardown: Clean up and close connections
	err := CloseMongoDB()
	if err != nil {
		fmt.Printf("Failed to close MongoDB connections after tests: %v", err)
	}

	// Clean up test log file
	os.Remove("./test.log")

	os.Exit(code)
}

// TestGetMongoClient tests the retrieval of a mongo client
func TestGetMongoClient(t *testing.T) {
	client, err := GetMongoClient(testDBInstance)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Test getting a non-existent client
	_, err = GetMongoClient("nonexistent")
	assert.Error(t, err)
}

// TestGetMongoDatabase tests the retrieval of a mongo database
func TestGetMongoDatabase(t *testing.T) {
	db, err := GetMongoDatabase(testDBInstance, testDatabaseName)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, testDatabaseName, db.Name())
}

// TestGetMongoCollection tests the retrieval of a mongo collection
func TestGetMongoCollection(t *testing.T) {
	coll, err := GetMongoCollection(testDBInstance, testDatabaseName, testCollection)
	assert.NoError(t, err)
	assert.NotNil(t, coll)
	assert.Equal(t, testCollection, coll.Name())
}

// User struct for testing CRUD operations
type testUser struct {
	ID   string `bson:"_id,omitempty"`
	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

// TestCRUDOperations tests create, read, update, and delete operations
func TestCRUDOperations(t *testing.T) {
	coll, err := GetMongoCollection(testDBInstance, testDatabaseName, testCollection)
	assert.NoError(t, err)

	// Clean up the collection before the test
	err = coll.Drop(context.Background())
	assert.NoError(t, err)

	// Create
	user := testUser{Name: "Alice", Age: 30}
	_, err = coll.InsertOne(context.Background(), user)
	assert.NoError(t, err)

	// Read
	var foundUser testUser
	err = coll.FindOne(context.Background(), bson.M{"name": "Alice"}).Decode(&foundUser)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", foundUser.Name)
	assert.Equal(t, 30, foundUser.Age)

	// Update
	update := bson.M{"$set": bson.M{"age": 31}}
	_, err = coll.UpdateOne(context.Background(), bson.M{"name": "Alice"}, update)
	assert.NoError(t, err)

	// Verify Update
	err = coll.FindOne(context.Background(), bson.M{"name": "Alice"}).Decode(&foundUser)
	assert.NoError(t, err)
	assert.Equal(t, 31, foundUser.Age)

	// Delete
	_, err = coll.DeleteOne(context.Background(), bson.M{"name": "Alice"})
	assert.NoError(t, err)

	// Verify Delete
	err = coll.FindOne(context.Background(), bson.M{"name": "Alice"}).Decode(&foundUser)
	assert.Error(t, err)
	assert.Equal(t, mongo.ErrNoDocuments, err)
}

// TestWithTransaction tests the transaction helper function
func TestWithTransaction(t *testing.T) {
	client, err := GetMongoClient(testDBInstance)
	assert.NoError(t, err)

	coll, err := GetMongoCollection(testDBInstance, testDatabaseName, testCollection)
	assert.NoError(t, err)

	// This test requires a MongoDB replica set. We will check for the error
	// and skip the test if transactions are not supported.
	err = WithTransaction(client, func(sessCtx mongo.SessionContext) error {
		_, err := coll.InsertOne(sessCtx, testUser{Name: "Bob", Age: 40})
		return err
	})

	if err != nil {
		// 检查是否是事务不支持的错误
		if cmdErr, ok := err.(mongo.CommandError); ok {
			// Error code 20: "Transaction numbers are only allowed on a replica set member or mongos"
			if cmdErr.Code == 20 {
				t.Skip("Skipping transaction test: MongoDB instance does not support transactions.")
			}
		}
		// 检查是否是写异常
		if writeException, ok := err.(mongo.WriteException); ok {
			isTransactionError := false
			for _, writeError := range writeException.WriteErrors {
				// Error code 20: "Transaction numbers are only allowed on a replica set member or mongos"
				if writeError.Code == 20 {
					isTransactionError = true
					break
				}
			}
			if isTransactionError {
				t.Skip("Skipping transaction test: MongoDB instance does not support transactions.")
			}
		}
		// 检查错误消息中是否包含事务相关的错误信息
		if err.Error() == "Transaction numbers are only allowed on a replica set member or mongos" {
			t.Skip("Skipping transaction test: MongoDB instance does not support transactions.")
		}
		// If it's another type of error, fail the test
		assert.NoError(t, err, "Transaction failed with an unexpected error")
	} else {
		// If transaction succeeded, verify the data and clean up
		var user testUser
		err = coll.FindOne(context.Background(), bson.M{"name": "Bob"}).Decode(&user)
		assert.NoError(t, err)
		assert.Equal(t, "Bob", user.Name)

		_, err = coll.DeleteOne(context.Background(), bson.M{"name": "Bob"})
		assert.NoError(t, err)
	}
}

// TestHealthCheck tests the health check function
func TestHealthCheck(t *testing.T) {
	healthStatus := HealthCheck()
	assert.NotEmpty(t, healthStatus)
	assert.Contains(t, healthStatus, testDBInstance)

	status := healthStatus[testDBInstance]
	assert.Equal(t, "success", status.Status)
	assert.Empty(t, status.Error)
	// 延迟时间可能为 0（非常快的连接），所以只检查它不为负数
	assert.GreaterOrEqual(t, status.Latency, int64(0))
	assert.GreaterOrEqual(t, status.MaxPool, 0)
	assert.GreaterOrEqual(t, status.InUse, 0)
	assert.GreaterOrEqual(t, status.Available, 0)
}

// TestParallelUsage tests the thread-safety of the connection pool
func TestParallelUsage(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coll, err := GetMongoCollection(testDBInstance, testDatabaseName, "parallel_test")
			assert.NoError(t, err)
			_, err = coll.InsertOne(context.Background(), testUser{Name: "Parallel", Age: i})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	// Clean up
	coll, err := GetMongoCollection(testDBInstance, testDatabaseName, "parallel_test")
	assert.NoError(t, err)
	err = coll.Drop(context.Background())
	assert.NoError(t, err)
}
