package mysql

import (
	"context"
	"errors"
	"time"

	gocommonlog "github.com/jessewkun/gocommon/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMysqlLogger 创建一个mysql日志记录器
func newMysqlLogger(slowThreshold time.Duration, level logger.LogLevel, ignore bool) *mysqlLogger {
	return &mysqlLogger{
		SlowThreshold:             slowThreshold,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: ignore,
	}
}

var _ logger.Interface = (*mysqlLogger)(nil)

func (ml *mysqlLogger) LogMode(lev logger.LogLevel) logger.Interface {
	newLogger := *ml
	newLogger.LogLevel = lev
	return &newLogger
}

func (ml *mysqlLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if ml.LogLevel >= logger.Info {
		gocommonlog.Info(ctx, TAG, msg, args...)
	}
}

func (ml *mysqlLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if ml.LogLevel >= logger.Warn {
		gocommonlog.Warn(ctx, TAG, msg, args...)
	}
}

func (ml *mysqlLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if ml.LogLevel >= logger.Error {
		gocommonlog.ErrorWithMsg(ctx, TAG, msg, args...)
	}
}

func (ml *mysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 错误日志：需要检查LogLevel >= Error
	if err != nil && ml.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !ml.IgnoreRecordNotFoundError) {
		gocommonlog.ErrorWithField(ctx, TAG, "MYSQL_QUERY_ERROR", map[string]interface{}{
			"sql":     sql,
			"rows":    rows,
			"elapsed": elapsed,
			"err":     err,
		})
		return
	}

	// 慢查询日志：需要检查LogLevel >= Warn
	if ml.SlowThreshold != 0 && elapsed > ml.SlowThreshold && ml.LogLevel >= logger.Warn {
		gocommonlog.WarnWithField(ctx, TAG, "MYSQL_SLOW_QUERY", map[string]interface{}{
			"sql":           sql,
			"rows":          rows,
			"elapsed":       elapsed,
			"slowthreshold": ml.SlowThreshold,
		})
		return
	}

	// 普通查询日志：需要检查LogLevel == Info（与GORM标准保持一致）
	if ml.LogLevel == logger.Info {
		gocommonlog.InfoWithField(ctx, TAG, "MYSQL_QUERY", map[string]interface{}{
			"sql":     sql,
			"rows":    rows,
			"elapsed": elapsed,
		})
	}
}
