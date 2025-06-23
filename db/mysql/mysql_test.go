package mysql

import (
	"os"
	"sync"
	"testing"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var mysqlTestMutex sync.Mutex

func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()
	code := m.Run()
	os.Remove("./test.log")
	os.Exit(code)
}

func TestInit(t *testing.T) {
	mysqlTestMutex.Lock()
	defer mysqlTestMutex.Unlock()

	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		Close()
	})

	testCases := []struct {
		name       string
		setupCfgs  func()
		checkError func(t *testing.T, err error)
		checkConns func(t *testing.T, initErr error)
	}{
		{
			name: "successful initialization",
			setupCfgs: func() {
				dsn := os.Getenv("TEST_MYSQL_DSN")
				if dsn == "" {
					dsn = "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
				}
				Cfgs = map[string]*Config{
					"test_db": {Dsn: []string{dsn}},
				}
			},
			checkError: func(t *testing.T, err error) {
				if err != nil {
					t.Logf("Init failed as expected without a running DB: %v", err)
				}
			},
			checkConns: func(t *testing.T, initErr error) {
				if initErr == nil {
					connList.mu.RLock()
					_, ok := connList.conns["test_db"]
					connList.mu.RUnlock()
					assert.True(t, ok, "connection for 'test_db' should have been created")
				} else {
					t.Log("Skipping connection check as Init failed")
				}
			},
		},
		{
			name: "initialization with no config",
			setupCfgs: func() {
				Cfgs = make(map[string]*Config)
			},
			checkError: func(t *testing.T, err error) {
				assert.NoError(t, err, "Init should not error with empty config")
			},
			checkConns: func(t *testing.T, initErr error) {
				connList.mu.RLock()
				defer connList.mu.RUnlock()
				assert.Empty(t, connList.conns, "connection list should be empty with no config")
			},
		},
		{
			name: "initialization with bad config (no dsn)",
			setupCfgs: func() {
				Cfgs = map[string]*Config{"bad_db": {}}
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err, "Init should error with bad config")
				assert.Contains(t, err.Error(), "mysql dsn is invalid")
			},
			checkConns: func(t *testing.T, initErr error) {
				connList.mu.RLock()
				defer connList.mu.RUnlock()
				assert.Empty(t, connList.conns, "connection list should be empty after a failed init")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Close()
			tc.setupCfgs()

			err := Init()

			if tc.checkError != nil {
				tc.checkError(t, err)
			}
			if tc.checkConns != nil {
				tc.checkConns(t, err)
			}
		})
	}
}

func TestGetConn_And_Close(t *testing.T) {
	mysqlTestMutex.Lock()
	defer mysqlTestMutex.Unlock()

	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		Close()
	})

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	}
	Cfgs = map[string]*Config{
		"test_cache": {Dsn: []string{dsn}},
	}
	if err := Init(); err != nil {
		t.Skipf("skipping test; could not connect to mysql: %v", err)
	}

	t.Run("get existing connection", func(t *testing.T) {
		conn, err := GetConn("test_cache")
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("get non-existent connection", func(t *testing.T) {
		_, err := GetConn("nonexistent")
		assert.Error(t, err)
	})

	t.Run("close connections", func(t *testing.T) {
		err := Close()
		assert.NoError(t, err)
		connList.mu.RLock()
		assert.Empty(t, connList.conns)
		connList.mu.RUnlock()
	})
}

func TestTransaction(t *testing.T) {
	mysqlTestMutex.Lock()
	defer mysqlTestMutex.Unlock()

	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		Close()
	})

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	}
	Cfgs = map[string]*Config{"tx_db": {Dsn: []string{dsn}}}
	if err := Init(); err != nil {
		t.Skipf("skipping transaction test; could not connect to mysql: %v", err)
	}

	db, err := GetConn("tx_db")
	assert.NoError(t, err)

	type TempUser struct {
		gorm.Model
		Name string
	}
	err = db.AutoMigrate(&TempUser{})
	assert.NoError(t, err)
	defer db.Migrator().DropTable(&TempUser{})

	t.Run("commit transaction", func(t *testing.T) {
		tx := NewTransaction(db)
		err := tx.tx.Create(&TempUser{Name: "commit_test"}).Error
		assert.NoError(t, err)
		assert.NoError(t, tx.Commit())

		var user TempUser
		res := db.Where("name = ?", "commit_test").First(&user)
		assert.NoError(t, res.Error)
		assert.Equal(t, "commit_test", user.Name)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		tx := NewTransaction(db)
		err := tx.tx.Create(&TempUser{Name: "rollback_test"}).Error
		assert.NoError(t, err)
		assert.NoError(t, tx.Rollback())

		var user TempUser
		res := db.Where("name = ?", "rollback_test").First(&user)
		assert.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)
	})
}
