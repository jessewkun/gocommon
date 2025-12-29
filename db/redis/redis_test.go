package redis

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
)

var redisTestMutex sync.Mutex

// TestMain manages setup for all tests in this package.
func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()
	// Run all tests
	code := m.Run()
	// Cleanup after all tests
	os.Remove("./test.log")
	os.Exit(code)
}

// TestNewManager tests the NewManager constructor function.
func TestNewManager(t *testing.T) {
	redisTestMutex.Lock()
	defer redisTestMutex.Unlock()

	testCases := []struct {
		name       string
		configs    map[string]*Config
		checkError func(t *testing.T, err error)
		checkMgr   func(t *testing.T, mgr *Manager)
	}{
		{
			name: "successful initialization single node",
			configs: map[string]*Config{
				"test_db": {Addrs: []string{"localhost:6379"}, IsCluster: false},
			},
			checkError: func(t *testing.T, err error) {
				if err != nil {
					t.Logf("NewManager failed as DB might not be running, which is acceptable: %v", err)
				}
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				conn, err := mgr.GetConn("test_db")
				// The connection should be in the map even if the ping failed
				assert.NotNil(t, conn)
				assert.NoError(t, err)
				assert.IsType(t, &redis.Client{}, conn)
			},
		},
		{
			name: "successful initialization cluster",
			configs: map[string]*Config{
				"test_cluster": {Addrs: []string{"localhost:6379"}, IsCluster: true},
			},
			checkError: func(t *testing.T, err error) {
				if err != nil {
					t.Logf("NewManager failed as DB might not be running, which is acceptable: %v", err)
				}
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				conn, err := mgr.GetConn("test_cluster")
				assert.NotNil(t, conn)
				assert.NoError(t, err)
				assert.IsType(t, &redis.ClusterClient{}, conn)
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
			name:    "initialization with bad config (no address)",
			configs: map[string]*Config{"bad_db": {}},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err, "NewManager should error with bad config")
				assert.Contains(t, err.Error(), "redis addrs is empty")
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				assert.Empty(t, mgr.conns, "connection map should be empty after a failed init")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure a clean state for each test case.
			if defaultManager != nil {
				_ = Close()
			}

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

// TestGlobalFunctions tests Init, GetConn, and Close global functions.
func TestGlobalFunctions(t *testing.T) {
	redisTestMutex.Lock()
	defer redisTestMutex.Unlock()

	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		if defaultManager != nil {
			_ = Close()
		}
	})

	Cfgs = map[string]*Config{
		"cache": {
			Addrs: []string{"localhost:6379"},
		},
	}
	if Init() != nil {
		t.Skip("skipping test; could not connect to redis. please ensure redis is running on localhost:6379")
	}

	t.Run("get existing connection", func(t *testing.T) {
		conn, errGet := GetConn("cache")
		assert.NoError(t, errGet)
		assert.NotNil(t, conn)

		pong, errPing := conn.Ping(context.Background()).Result()
		assert.NoError(t, errPing)
		assert.Equal(t, "PONG", pong)
	})

	t.Run("get non-existent connection", func(t *testing.T) {
		_, errGet := GetConn("nonexistent")
		assert.Error(t, errGet)
	})

	t.Run("close connections", func(t *testing.T) {
		errClose := Close()
		assert.NoError(t, errClose)
		assert.NotNil(t, defaultManager)
		defaultManager.mu.RLock()
		assert.Empty(t, defaultManager.conns)
		defaultManager.mu.RUnlock()
	})
}

func TestHealthCheck(t *testing.T) {
	// This is a basic test. A more thorough test would set up connections first.
	healthStatus := HealthCheck()
	assert.NotNil(t, healthStatus)
}
