package cache

import (
	"context"
	"time"

	"github.com/jessewkun/gocommon/logger"

	"github.com/go-redis/redis/v8"
)

type RedisHook struct{}

func (h *RedisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "redis_start_time", time.Now())
	return ctx, nil
}

func (h *RedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	startTime := ctx.Value("redis_start_time").(time.Time)
	duration := time.Since(startTime)
	logger.InfoWithField(ctx, TAGNAME, "AfterProcess", map[string]interface{}{
		"cmd":      cmd,
		"duration": duration,
	})
	return nil
}

func (h *RedisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "redis_start_time", time.Now())
	return ctx, nil
}

func (h *RedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	startTime := ctx.Value("redis_start_time").(time.Time)
	duration := time.Since(startTime)
	logger.InfoWithField(ctx, TAGNAME, "AfterProcessPipeline", map[string]interface{}{
		"cmd":      cmds,
		"duration": duration,
	})
	return nil
}
