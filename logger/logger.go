// Package logger 提供日志记录功能
package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jessewkun/gocommon/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logzap   *zap.Logger
	hostname string
	localIP  string
	once     sync.Once
)

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	PanicLevel LogLevel = "panic"
	FatalLevel LogLevel = "fatal"
)

// LogEntry 日志条目
type LogEntry struct {
	Level   LogLevel
	Tag     string
	Message string
	Fields  map[string]interface{}
	Error   error
}

// 初始化系统信息
func initSystemInfo() {
	once.Do(func() {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			hostname = "unknown"
			fmt.Printf("Failed to get hostname: %v\n", err)
		}
		localIP, err = utils.GetLocalIP()
		if err != nil {
			localIP = "unknown"
			fmt.Printf("Failed to get local IP: %v\n", err)
		}
	})
}

func Zap() *zap.Logger {
	return logzap
}

func Init() error {
	if err := Cfg.Validate(); err != nil {
		return fmt.Errorf("invalid logger config: %v", err)
	}

	initSystemInfo()
	logzap = zap.New(initCore(), zap.AddCallerSkip(2), zap.AddCaller())
	return nil
}

func initCore() zapcore.Core {
	opts := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   Cfg.Path,      // ⽇志⽂件路径
			MaxSize:    Cfg.MaxSize,   // 单位为MB,默认为100MB
			MaxAge:     Cfg.MaxAge,    // 文件最多保存多少天
			LocalTime:  true,          // 采用本地时间
			Compress:   false,         // 是否压缩日志
			MaxBackups: Cfg.MaxBackup, // 保留旧文件的最大个数
		}),
	}

	syncWriter := zapcore.NewMultiWriteSyncer(opts...)
	logLevel, err := zapcore.ParseLevel(Cfg.LogLevel)
	if err != nil {
		logLevel = zapcore.DebugLevel
	}

	encoderConf := zapcore.EncoderConfig{
		CallerKey:     "caller_line",
		LevelKey:      "level",
		MessageKey:    "msg",
		TimeKey:       "datetime",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}, // 自定义时间格式
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 日志级别改为大写
		EncodeCaller:   zapcore.FullCallerEncoder,   // 全路径编码器
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConf),
		syncWriter, zap.NewAtomicLevelAt(logLevel))
}

// formatField 格式化字段
func formatField(c context.Context, tag string) []zapcore.Field {
	fields := make([]zapcore.Field, 0)

	fields = append(fields, zap.String("tag", tag))
	fields = append(fields, zap.String("host", hostname))
	fields = append(fields, zap.String("ip", localIP))

	// 日志强制添加 trace_id 和 user_id
	fields = append(fields, FieldsFromCtx(c)...)

	if len(Cfg.TransparentParameter) > 0 {
		for _, v := range Cfg.TransparentParameter {
			if value := c.Value(v); value != nil {
				fields = append(fields, zap.Any(v, value))
			}
		}
	}

	return fields
}

// log 统一的日志记录函数
func log(c context.Context, entry LogEntry) {
	if Cfg.Closed {
		return
	}

	fields := formatField(c, entry.Tag)

	// 添加自定义字段
	for k, v := range entry.Fields {
		fields = append(fields, zap.Any(k, v))
	}

	// 添加错误信息
	if entry.Error != nil {
		fields = append(fields, zap.Error(entry.Error))
	}

	// 记录日志
	switch entry.Level {
	case DebugLevel:
		logzap.Debug(entry.Message, fields...)
	case InfoLevel:
		logzap.Info(entry.Message, fields...)
	case WarnLevel:
		logzap.Warn(entry.Message, fields...)
	case ErrorLevel:
		logzap.Error(entry.Message, fields...)
	case PanicLevel:
		logzap.Panic(entry.Message, fields...)
	case FatalLevel:
		logzap.Fatal(entry.Message, fields...)
	}

	// 发送报警
	SendAlarm(c, string(entry.Level), entry.Tag, entry.Message, entry.Error)
}

// Info 记录信息日志
func Info(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   InfoLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}

// InfoWithField 记录带字段的信息日志
func InfoWithField(c context.Context, tag string, msg string, fields map[string]interface{}) {
	log(c, LogEntry{
		Level:   InfoLevel,
		Tag:     tag,
		Message: msg,
		Fields:  fields,
	})
}

// InfoWithAlarm 记录带报警的信息日志
func InfoWithAlarm(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   InfoLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
	ForceSendAlarm(c, string(InfoLevel), tag, fmt.Sprintf(msg, args...), nil)
}

// Error 记录错误日志
func Error(c context.Context, tag string, err error) {
	log(c, LogEntry{
		Level:   ErrorLevel,
		Tag:     tag,
		Message: "",
		Error:   err,
	})
}

// ErrorWithMsg 记录带消息的错误日志
func ErrorWithMsg(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   ErrorLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}

// ErrorWithField 记录带字段的错误日志
func ErrorWithField(c context.Context, tag string, msg string, fields map[string]interface{}) {
	log(c, LogEntry{
		Level:   ErrorLevel,
		Tag:     tag,
		Message: msg,
		Fields:  fields,
	})
}

// Debug 记录调试日志
func Debug(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   DebugLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}

// DebugWithField 记录带字段的调试日志
func DebugWithField(c context.Context, tag string, msg string, fields map[string]interface{}) {
	log(c, LogEntry{
		Level:   DebugLevel,
		Tag:     tag,
		Message: msg,
		Fields:  fields,
	})
}

// Warn 记录警告日志
func Warn(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   WarnLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}

// WarnWithError 记录带错误信息的警告日志
func WarnWithError(c context.Context, tag string, err error) {
	log(c, LogEntry{
		Level:   WarnLevel,
		Tag:     tag,
		Message: "",
		Error:   err,
	})
}

// WarnWithField 记录带字段的警告日志
func WarnWithField(c context.Context, tag string, msg string, fields map[string]interface{}) {
	log(c, LogEntry{
		Level:   WarnLevel,
		Tag:     tag,
		Message: msg,
		Fields:  fields,
	})
}

// Panic 记录紧急日志
func Panic(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   PanicLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}

// Fatal 记录致命日志
func Fatal(c context.Context, tag string, msg string, args ...interface{}) {
	log(c, LogEntry{
		Level:   FatalLevel,
		Tag:     tag,
		Message: fmt.Sprintf(msg, args...),
	})
}
