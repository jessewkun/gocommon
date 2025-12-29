package nacos

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var nacosTestMutex sync.Mutex

// TestNewManager tests the NewManager constructor function.
func TestNewManager(t *testing.T) {
	nacosTestMutex.Lock()
	defer nacosTestMutex.Unlock()

	testCases := []struct {
		name       string
		configs    map[string]*Config
		checkError func(t *testing.T, err error)
		checkMgr   func(t *testing.T, mgr *Manager)
	}{
		{
			name: "successful initialization",
			configs: map[string]*Config{
				"test_nacos": {Host: "localhost", Port: 8848},
			},
			checkError: func(t *testing.T, err error) {
				if err != nil {
					// In a CI/CD environment, Nacos might not be running.
					// We accept a connection error but ensure the manager is still created.
					t.Logf("NewManager failed as Nacos might not be running, which is acceptable for this test: %v", err)
				}
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				client, err := mgr.GetClient("test_nacos")
				assert.NoError(t, err)
				assert.NotNil(t, client)
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
				assert.Empty(t, mgr.clients, "clients map should be empty with no config")
			},
		},
		{
			name: "initialization with bad config (e.g. empty host)",
			configs: map[string]*Config{
				"bad_nacos": {Port: 8848}, // Host is missing, will be defaulted
			},
			checkError: func(t *testing.T, err error) {
				if err != nil {
					t.Logf("NewManager failed as Nacos might not be running: %v", err)
				}
			},
			checkMgr: func(t *testing.T, mgr *Manager) {
				assert.NotNil(t, mgr)
				_, err := mgr.GetClient("bad_nacos")
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr, err := NewManager(tc.configs)
			if mgr != nil {
				defer func() { _ = mgr.Close() }()
			}

			if tc.checkError != nil {
				tc.checkError(t, err)
			}
			if tc.checkMgr != nil {
				tc.checkMgr(t, mgr)
			}
		})
	}
}

// TestGlobalFunctions tests Init, GetClient, and Close global functions.
func TestGlobalFunctions(t *testing.T) {
	nacosTestMutex.Lock()
	defer nacosTestMutex.Unlock()

	// Save original global state and restore it after the test.
	originalCfgs := Cfgs
	t.Cleanup(func() {
		Cfgs = originalCfgs
		if defaultManager != nil {
			_ = Close()
			defaultManager = nil
		}
	})

	// Test before Init
	t.Run("access before init", func(t *testing.T) {
		_, err := GetClient("any")
		assert.Error(t, err)
		assert.Equal(t, "nacos manager is not initialized", err.Error())

		err = Close()
		assert.Error(t, err)
		assert.Equal(t, "nacos manager is not initialized", err.Error())
	})

	Cfgs = map[string]*Config{
		"default": {
			Host: "localhost",
			Port: 8848,
		},
	}
	err := Init()
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		// If Nacos is not running, we might get a timeout or connection error.
		// We can skip the rest of the test in that case.
		t.Skipf("skipping test; could not connect to Nacos. please ensure Nacos is running on localhost:8848. error: %v", err)
		return
	}

	t.Run("get existing client", func(t *testing.T) {
		client, errGet := GetClient("default")
		assert.NoError(t, errGet)
		assert.NotNil(t, client)
	})

	t.Run("get non-existent client", func(t *testing.T) {
		_, errGet := GetClient("nonexistent")
		assert.Error(t, errGet)
		assert.Contains(t, errGet.Error(), "nacos client 'nonexistent' not found")
	})

	t.Run("close connections", func(t *testing.T) {
		errClose := Close()
		assert.NoError(t, errClose)
		assert.NotNil(t, defaultManager)
		defaultManager.mu.RLock()
		assert.Empty(t, defaultManager.clients)
		defaultManager.mu.RUnlock()

		// Verify manager is cleared
		defaultManager = nil
		_, errAfterClose := GetClient("default")
		assert.Error(t, errAfterClose, "should fail after manager is cleared")
	})
}
