package mysql

import (
	"os"
	"testing"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
)

// TestMain sets up a temporary logger for tests.
func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()
	code := m.Run()
	os.Remove("./test.log")
	os.Exit(code)
}

// TestNewManager tests the NewManager constructor function.
func TestNewManager(t *testing.T) {
	testCases := []struct {
		name       string
		configs    map[string]*Config
		checkError func(t *testing.T, err error)
		checkMgr   func(t *testing.T, mgr *Manager)
	}{
		{
			name: "successful initialization",
			configs: (func() map[string]*Config {
				dsn := "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
				return map[string]*Config{
					"test_db": {Dsn: []string{dsn}},
				}
			})(),
			checkError: func(t *testing.T, err error) {
				if err != nil {
					t.Logf("NewManager failed as DB might not be running, which is acceptable for this test path: %v", err)
				}
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				// If manager was created, check if the connection exists, even if it failed to connect.
				// The manager holds the config regardless of connection success.
				mgr.mu.RLock()
				_, ok := mgr.conns["test_db"]
				mgr.mu.RUnlock()
				assert.True(t, ok, "connection config for 'test_db' should have been processed")
			},
		},
		{
			name:    "initialization with no config",
			configs: make(map[string]*Config),
			checkError: func(t *testing.T, err error) {
				assert.NoError(t, err, "NewManager should not error with empty config")
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				assert.Empty(t, mgr.conns, "connection map should be empty with no config")
			},
		},
		{
			name:    "initialization with bad config (no dsn)",
			configs: map[string]*Config{"bad_db": {Dsn: []string{}}},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err, "NewManager should error with bad config")
				assert.Contains(t, err.Error(), "mysql dsn is invalid")
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				assert.Empty(t, mgr.conns, "connection map should be empty after a failed init")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr, err := NewManager(tc.configs)
			defer func() {
				if mgr != nil {
					_ = mgr.Close()
				}
			}()

			if tc.checkError != nil {
				tc.checkError(t, err)
			}
			if tc.checkMgr != nil {
				tc.checkMgr(t, mgr)
			}
		})
	}
}

// TestGetConn_And_Close tests the GetConn and Close methods of the Manager.
func TestGetConn_And_Close(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	configs := map[string]*Config{
		"test_cache": {Dsn: []string{dsn}},
	}

	mgr, err := NewManager(configs)
	if err != nil {
		t.Skipf("skipping test; could not connect to mysql: %v", err)
	}
	defer mgr.Close()

	t.Run("get existing connection", func(t *testing.T) {
		conn, err := mgr.GetConn("test_cache")
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("get non-existent connection", func(t *testing.T) {
		_, err := mgr.GetConn("nonexistent")
		assert.Error(t, err)
	})

	t.Run("close connections", func(t *testing.T) {
		errs := mgr.Close()
		assert.NoError(t, errs)
		mgr.mu.RLock()
		assert.Empty(t, mgr.conns)
		mgr.mu.RUnlock()
	})
}
