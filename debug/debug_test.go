package debug

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jessewkun/gocommon/logger"
)

// debugTestMutex ensures that tests modifying the global Cfg are run serially.
var debugTestMutex sync.Mutex

func TestIsDebug(t *testing.T) {
	debugTestMutex.Lock()
	defer debugTestMutex.Unlock()

	originalModule := Cfg.Module
	t.Cleanup(func() { Cfg.Module = originalModule })

	Cfg.Module = []string{"mod1", "mod2"}
	if !IsDebug("mod1") {
		t.Error("mod1 应该被识别为 debug")
	}
	if IsDebug("mod3") {
		t.Error("mod3 不应该被识别为 debug")
	}
}

func TestInitDebug(t *testing.T) {
	debugTestMutex.Lock()
	defer debugTestMutex.Unlock()

	originalModule := Cfg.Module
	t.Cleanup(func() { Cfg.Module = originalModule })

	Cfg.Module = []string{"modx"}
	// hookPrint 只会在 IsDebug 返回 true 时调用
	f := InitDebug("modx")
	// 重定向 os.Stdout 捕获输出
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f(context.Background(), "hello %s", []interface{}{"world"}...)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("InitDebug 未输出预期内容, got: %s", output)
	}
}

func TestHookPrint_ModeLog(t *testing.T) {
	debugTestMutex.Lock()
	defer debugTestMutex.Unlock()

	originalMode := Cfg.Mode
	t.Cleanup(func() { Cfg.Mode = originalMode })

	Cfg.Mode = "log"
	// 为了测试日志输出，我们需要临时初始化logger
	// 注意：这可能会与其他测试（如http_test）冲突，因此锁是必要的
	originalLogCfg := logger.Cfg
	logger.Cfg.Path = "./test.log"
	logger.Cfg.Closed = false
	if err := logger.Init(); err != nil {
		t.Fatalf("logger.InitLogger() failed: %v", err)
	}
	t.Cleanup(func() {
		logger.Cfg = originalLogCfg
		logger.Init()
		os.Remove("./test.log")
	})

	hookPrint(context.Background(), "log模式: %d", 123)
	// 检查日志文件是否包含内容
	content, err := os.ReadFile("./test.log")
	if err != nil {
		t.Fatalf("Could not read log file: %v", err)
	}
	if !strings.Contains(string(content), "log模式: 123") {
		t.Errorf("log file should contain the message, but got: %s", string(content))
	}
}

func TestHookPrint_ModeStdout(t *testing.T) {
	debugTestMutex.Lock()
	defer debugTestMutex.Unlock()

	originalMode := Cfg.Mode
	t.Cleanup(func() { Cfg.Mode = originalMode })

	Cfg.Mode = "console"
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	hookPrint(context.Background(), "stdout模式: %s", "abc")
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	if !strings.Contains(output, "stdout模式: abc") {
		t.Errorf("hookPrint 未输出预期内容, got: %s", output)
	}
}
