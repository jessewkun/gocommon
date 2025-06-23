package safego

import (
	"context"
	"os"
	"testing"

	"github.com/jessewkun/gocommon/logger"
)

func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()
	// Run all tests
	code := m.Run()
	// Cleanup after all tests
	os.Remove("./test.log")
	os.Exit(code)
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
