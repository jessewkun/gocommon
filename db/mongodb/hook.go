package mongodb

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/logger"
	"go.mongodb.org/mongo-driver/event"
)

// commandMonitor 实现了 event.CommandMonitor 接口
type commandMonitor struct {
	slowThreshold time.Duration
	reqMap        sync.Map // map[int64]time.Time
}

// newCommandMonitor 创建一个新的监控器
func newCommandMonitor(slowThreshold time.Duration) *commandMonitor {
	if slowThreshold <= 0 {
		slowThreshold = 500 * time.Millisecond // 默认值
	}
	return &commandMonitor{
		slowThreshold: slowThreshold,
	}
}

// Started 在命令开始时调用
func (m *commandMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	m.reqMap.Store(evt.RequestID, time.Now())
}

// Succeeded 在命令成功时调用
func (m *commandMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	m.logCommand(ctx, evt.CommandFinishedEvent, nil)
}

// Failed 在命令失败时调用
func (m *commandMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	m.logCommand(ctx, evt.CommandFinishedEvent, errors.New(evt.Failure))
}

func (m *commandMonitor) logCommand(ctx context.Context, evt event.CommandFinishedEvent, err error) {
	if _, ok := m.reqMap.LoadAndDelete(evt.RequestID); !ok {
		return
	}

	duration := evt.Duration
	fields := map[string]interface{}{
		"cmd":      evt.CommandName,
		"duration": duration,
		"req_id":   evt.RequestID,
		"status":   "success",
	}

	if err != nil {
		fields["status"] = "error"
		fields["error"] = err.Error()
	}

	if duration > m.slowThreshold {
		logger.WarnWithField(ctx, TAG, "MONGO_SLOW_QUERY", fields)
	} else {
		logger.InfoWithField(ctx, TAG, "MONGO_QUERY", fields)
	}
}
