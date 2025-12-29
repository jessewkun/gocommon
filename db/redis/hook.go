package redis

import (
	"context"
	"time"

	"github.com/jessewkun/gocommon/logger"

	"github.com/go-redis/redis/v8"
)

// RedisContextKey 定义 Redis context key 的类型
type RedisContextKey string

const (
	// RedisStartTimeKey Redis 开始时间 key
	RedisStartTimeKey RedisContextKey = "redis_start_time"
)

// RedisHook Redis钩子
type RedisHook struct {
	slowThreshold time.Duration // 慢查询阈值
}

// NewRedisHook 创建Redis钩子
func newRedisHook(slowThreshold time.Duration) *RedisHook {
	if slowThreshold == 0 {
		slowThreshold = 100 * time.Millisecond
	}
	return &RedisHook{
		slowThreshold: slowThreshold,
	}
}

func (h *RedisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, RedisStartTimeKey, time.Now())
	return ctx, nil
}

func (h *RedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	startTimeVal := ctx.Value(RedisStartTimeKey)
	startTime, ok := startTimeVal.(time.Time)
	if !ok {
		logger.Warn(ctx, TAG, "Could not find start time in context for redis command")
		return nil
	}
	duration := time.Since(startTime)

	fields := map[string]interface{}{
		"cmd":      cmd.Args(),
		"duration": duration,
		"status":   "success",
	}

	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		fields["status"] = "error"
		fields["error"] = cmd.Err().Error()
	}

	// 记录慢查询
	if duration > h.slowThreshold {
		logger.WarnWithField(ctx, TAG, "REDIS_SLOW_QUERY", fields)
	} else {
		logger.InfoWithField(ctx, TAG, "REDIS_QUERY", fields)
	}

	return nil
}

func (h *RedisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, RedisStartTimeKey, time.Now())
	return ctx, nil
}

func (h *RedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	startTimeVal := ctx.Value(RedisStartTimeKey)
	startTime, ok := startTimeVal.(time.Time)
	if !ok {
		logger.Warn(ctx, TAG, "Could not find start time in context for redis command")
		return nil
	}
	duration := time.Since(startTime)

	// 统计成功和失败的命令数
	successCount := 0
	errorCount := 0
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	fields := map[string]interface{}{
		"cmd_count":     len(cmds),
		"success_count": successCount,
		"error_count":   errorCount,
		"duration":      duration,
		"status":        "success",
	}

	if errorCount > 0 {
		fields["status"] = "partial_error"
	}

	// 记录慢查询
	if duration > h.slowThreshold {
		logger.WarnWithField(ctx, TAG, "REDIS_SLOW_PIPELINE", fields)
	} else {
		logger.InfoWithField(ctx, TAG, "REDIS_PIPELINE", fields)
	}

	return nil
}
