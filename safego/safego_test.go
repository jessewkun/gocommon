package safego

import (
	"context"
	"testing"

	"github.com/jessewkun/gocommon/logger"
)

func init() {
	cfg := logger.DefaultConfig()
	cfg.Path = "./test.log"
	cfg.MaxSize = 1
	cfg.MaxAge = 1
	cfg.MaxBackup = 1
	cfg.AlarmLevel = "warn"
	_ = logger.InitLogger(cfg)
}

func TestWaitGroupWrapper_Wrap(t *testing.T) {
	var wg WaitGroupWrapper
	var called bool
	ctx := context.Background()
	wg.Wrap(ctx, func() {
		called = true
	})
	wg.Wait()
	if !called {
		t.Error("Wrap 未正常执行传入函数")
	}
}

func TestSafeGo_Normal(t *testing.T) {
	var called bool
	SafeGo(context.Background(), func() {
		called = true
	})
	if !called {
		t.Error("SafeGo 未正常执行传入函数")
	}
}

func TestSafeGo_Panic(t *testing.T) {
	// 只需保证 panic 被 recover，不会导致测试崩溃
	SafeGo(context.Background(), func() {
		panic("test panic")
	})
}
