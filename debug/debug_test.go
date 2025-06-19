package debug

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jessewkun/gocommon/logger"
	"github.com/spf13/viper"
)

func TestIsDebug(t *testing.T) {
	viper.Set("debug.module", []string{"mod1", "mod2"})
	if !IsDebug("mod1") {
		t.Error("mod1 应该被识别为 debug")
	}
	if IsDebug("mod3") {
		t.Error("mod3 不应该被识别为 debug")
	}
}

func TestInitDebug(t *testing.T) {
	viper.Set("debug.module", []string{"modx"})
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
	viper.Set("debug.mode", "log")
	// 初始化 logger，避免 logzap 为 nil
	cfg := logger.DefaultConfig()
	cfg.Path = "./test.log"
	cfg.MaxSize = 1
	cfg.MaxAge = 1
	cfg.MaxBackup = 1
	cfg.AlarmLevel = "warn"
	_ = logger.InitLogger(cfg)
	// logger.Debug 实际调用不可控，这里只测试分支覆盖
	// 只需保证不会 panic
	hookPrint(context.Background(), "log模式: %d", 123)
}

func TestHookPrint_ModeStdout(t *testing.T) {
	viper.Set("debug.mode", "")
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
