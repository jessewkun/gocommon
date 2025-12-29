package debug

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setup a helper to temporarily redirect stdout
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestIsDebug(t *testing.T) {
	// Clean up global state after test
	originalConfig := defaultDebugger.Config
	originalModules := defaultDebugger.enabledModules
	defer func() {
		defaultDebugger.mu.Lock()
		defaultDebugger.Config = originalConfig
		defaultDebugger.enabledModules = originalModules
		defaultDebugger.mu.Unlock()
	}()

	// Simulate loading a config
	v := viper.New()
	v.Set("debug.module", []string{"mod1", "mod2"})
	v.Set("debug.mode", "console")
	if err := defaultDebugger.Reload(v); err != nil {
		t.Fatalf("failed to reload debug config: %v", err)
	}

	assert.True(t, IsDebug("mod1"), "mod1 should be enabled for debug")
	assert.True(t, IsDebug("mod2"), "mod2 should be enabled for debug")
	assert.False(t, IsDebug("mod3"), "mod3 should not be enabled for debug")
}

func TestLog_ConsoleMode(t *testing.T) {
	// Clean up global state after test
	originalConfig := defaultDebugger.Config
	originalModules := defaultDebugger.enabledModules
	defer func() {
		defaultDebugger.mu.Lock()
		defaultDebugger.Config = originalConfig
		defaultDebugger.enabledModules = originalModules
		defaultDebugger.mu.Unlock()
	}()

	// Simulate loading config for console mode
	v := viper.New()
	v.Set("debug.module", []string{"test_console"})
	v.Set("debug.mode", "console")
	if err := defaultDebugger.Reload(v); err != nil {
		t.Fatalf("failed to reload debug config: %v", err)
	}

	// Case 1: Debug module is enabled
	output := captureStdout(func() {
		Log(context.Background(), "test_console", "hello %s", "world")
	})
	assert.Contains(t, output, "[DEBUG_test_console]", "Output should contain correct tag")
	assert.Contains(t, output, "hello world", "Output should contain the formatted message")

	// Case 2: Debug module is disabled
	output = captureStdout(func() {
		Log(context.Background(), "other_module", "this should not be printed")
	})
	assert.Empty(t, output, "There should be no output for a disabled module")
}
