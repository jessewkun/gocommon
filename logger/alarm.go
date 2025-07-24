package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/jessewkun/gocommon/alarm"
	"github.com/jessewkun/gocommon/config"
	"github.com/jessewkun/gocommon/constant"
)

// 报警级别映射
var alarmLevelMap = map[string][]string{
	"debug": {"debug", "info", "warn", "error", "fatal", "panic"},
	"info":  {"info", "warn", "error", "fatal", "panic"},
	"warn":  {"warn", "error", "fatal", "panic"},
	"error": {"error", "fatal", "panic"},
	"fatal": {"fatal", "panic"},
	"panic": {"panic"},
}

// SendAlarm 发送报警
func SendAlarm(c context.Context, level string, tag string, msg string, err error) {
	if Cfg.Closed {
		return
	}

	if Cfg.AlarmLevel == "" {
		return
	}

	allowedLevels, exists := alarmLevelMap[Cfg.AlarmLevel]
	if !exists {
		return
	}

	// 检查当前级别是否允许报警
	canAlarm := false
	for _, v := range allowedLevels {
		if v == level {
			canAlarm = true
			break
		}
	}

	if canAlarm {
		content := buildAlarmContent(c, tag, msg, err)
		_ = alarm.SendAlarm(c, "["+config.Cfg.Mode+"] Service Alarm - "+level, content)
	}
}

// ForceSendAlarm 强制发送报警
func ForceSendAlarm(c context.Context, level string, tag string, msg string, err error) {
	content := buildAlarmContent(c, tag, msg, err)
	_ = alarm.SendAlarm(c, "["+config.Cfg.Mode+"] Service Alarm - "+level, content)
}

// buildAlarmContent 构建报警内容
func buildAlarmContent(c context.Context, tag string, msg string, err error) []string {
	content := make([]string, 0)

	content = append(content, fmt.Sprintf("【DATETIME】: %s", time.Now().Format("2006-01-02 15:04:05")))
	content = append(content, fmt.Sprintf("【SERVER IP】: %s", localIP))
	content = append(content, fmt.Sprintf("【HOSTNAME】: %s", hostname))
	content = append(content, fmt.Sprintf("【TAG】: %s", tag))

	if msg != "" {
		content = append(content, fmt.Sprintf("【MESSAGE】: %s", msg))
	}

	if err != nil {
		content = append(content, fmt.Sprintf("【ERROR】: %v", err))
	}

	if stackInfo := getCallStackInfo(); stackInfo != "" {
		content = append(content, fmt.Sprintf("【CALL STACK】: %s", stackInfo))
	}

	if requestPath := c.Value(constant.CtxCurrentRequestPath); requestPath != nil {
		content = append(content, fmt.Sprintf("【REQUEST PATH】: %v", requestPath))
	}

	if userID := c.Value(constant.CtxUserID); userID != nil {
		content = append(content, fmt.Sprintf("【USER ID】: %v", userID))
	}

	if traceID := c.Value(constant.CtxTraceID); traceID != nil {
		content = append(content, fmt.Sprintf("【TRACE ID】: %v", traceID))
	}

	return content
}

// getCallStackInfo 获取调用栈信息
func getCallStackInfo() string {
	// 跳过当前函数和调用它的函数，获取实际的调用位置
	pc := make([]uintptr, 10)
	n := runtime.Callers(4, pc) // 跳过runtime.Callers, getCallStackInfo, buildAlarmContent, SendAlarm
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pc[:n])
	var stackInfo strings.Builder

	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		// 只获取第一个有效的调用帧（跳过logger包内部的调用）
		if !strings.Contains(frame.File, "logger/") {
			stackInfo.WriteString(fmt.Sprintf("%s:%d", frame.File, frame.Line))
			break
		}
	}

	return stackInfo.String()
}
