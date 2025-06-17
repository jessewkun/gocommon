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
	return &mysqlLogger{}
}

func (ml *mysqlLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	gocommonlog.Info(ctx, TAGNAME, msg, args...)
}

func (ml *mysqlLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	gocommonlog.Warn(ctx, TAGNAME, msg, args...)
}

func (ml *mysqlLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	gocommonlog.ErrorWithMsg(ctx, TAGNAME, msg, args...)
}

func (ml *mysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !ml.IgnoreRecordNotFoundError) {
		gocommonlog.ErrorWithField(ctx, TAGNAME, "MYSQL_QUERY_ERROR", map[string]interface{}{
			"sql":     sql,
			"rows":    rows,
			"elapsed": elapsed,
			"err":     err,
		})
		return
	}

	if ml.SlowThreshold != 0 && elapsed > ml.SlowThreshold {
		gocommonlog.WarnWithField(ctx, TAGNAME, "MYSQL_SLOW_QUERY", map[string]interface{}{
			"sql":           sql,
			"rows":          rows,
			"elapsed":       elapsed,
			"slowthreshold": ml.SlowThreshold,
		})
	} else {
		gocommonlog.InfoWithField(ctx, TAGNAME, "MYSQL_QUERY", map[string]interface{}{
			"sql":     sql,
			"rows":    rows,
			"elapsed": elapsed,
		})
	}
}
