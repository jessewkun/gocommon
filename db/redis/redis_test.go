package redis

import (
	"context"
	"os"
	"sync"
	"testing"

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

func TestInit(t *testing.T) {
	redisTestMutex.Lock()
	defer redisTestMutex.Unlock()

	// Save and restore original global config for the entire test function.
	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		Close()
	})

	testCases := []struct {
		name       string
		setupCfgs  func()
		checkError func(t *testing.T, err error)
		checkConns func(t *testing.T)
	}{
		{
			name: "successful initialization",
			setupCfgs: func() {
				Cfgs = map[string]*Config{
					"test_db": {Addrs: []string{"localhost:6379"}},
				}
			},
			checkError: func(t *testing.T, err error) {
				// We don't assert a specific error since it depends on a live redis.
				// An error is acceptable if redis is not running.
			},
			checkConns: func(t *testing.T) {
				connList.mu.RLock()
				defer connList.mu.RUnlock()
				_, ok := connList.conns["test_db"]
				assert.True(t, ok, "connection for 'test_db' should have been created")
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
			checkConns: func(t *testing.T) {
				connList.mu.RLock()
				defer connList.mu.RUnlock()
				assert.Empty(t, connList.conns, "connection list should be empty with no config")
			},
		},
		{
			name: "initialization with bad config",
			setupCfgs: func() {
				Cfgs = map[string]*Config{
					"bad_db": {}, // Config with no address
				}
			},
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err, "Init should error with bad config")
				assert.Contains(t, err.Error(), "redis addrs is empty")
			},
			checkConns: func(t *testing.T) {
				connList.mu.RLock()
				defer connList.mu.RUnlock()
				assert.Empty(t, connList.conns, "connection list should be empty after a failed init")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure a clean state for each test case.
			Close()
			tc.setupCfgs()

			err := Init()

			if tc.checkError != nil {
				tc.checkError(t, err)
			}
			if tc.checkConns != nil {
				tc.checkConns(t)
			}
		})
	}
}

func TestGetConn_And_Close(t *testing.T) {
	redisTestMutex.Lock()
	defer redisTestMutex.Unlock()

	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		Close()
	})

	Cfgs = map[string]*Config{
		"cache": {
			Addrs: []string{"localhost:6379"},
		},
	}
	// This part of the test requires a running Redis server.
	// We call Init and check for an error to see if we can proceed.
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
		connList.mu.RLock()
		assert.Empty(t, connList.conns)
		connList.mu.RUnlock()
	})
}

func TestHealthCheck(t *testing.T) {
	// This is a basic test. A more thorough test would set up connections first.
	healthStatus := HealthCheck()
	assert.NotNil(t, healthStatus)
}
