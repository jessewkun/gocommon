package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

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

// RegisterAlerter 注册一个报警器实例，用于发送报警
func RegisterAlerter(a Alerter) {
	if a != nil {
		alerter = a
	}
}

// SendAlarm 根据配置的级别发送报警
func SendAlarm(c context.Context, level string, tag string, msg string, err error) {
	if Cfg.Closed || Cfg.AlarmLevel == "" {
		return
	}

	allowedLevels, exists := alarmLevelMap[Cfg.AlarmLevel]
	if !exists {
		return
	}

	canAlarm := false
	for _, v := range allowedLevels {
		if v == level {
			canAlarm = true
			break
		}
	}

	if canAlarm {
		ForceSendAlarm(c, level, tag, msg, err)
	}
}

// ForceSendAlarm 强制发送报警，忽略级别配置
func ForceSendAlarm(c context.Context, level string, tag string, msg string, err error) {
	if alerter == nil {
		// 如果没有注册报警器，只打印一条日志提示
		log(c, LogEntry{
			Level:   WarnLevel,
			Tag:     "ALERTER_NOT_REGISTERED",
			Message: "Alerter is not registered, unable to send alarm.",
			Fields: map[string]interface{}{
				"origin_level": level,
				"origin_tag":   tag,
				"origin_msg":   msg,
				"origin_err":   err,
			},
		})
		return
	}

	title := fmt.Sprintf("[%s] %s Alarm - %s", config.Cfg.Mode, config.Cfg.AppName, level)
	content := buildAlarmContent(c, tag, msg, err)

	if sendErr := alerter.Send(c, title, content); sendErr != nil {
		// 报警发送失败，记录错误日志，但不再触发新的报警，避免循环
		log(c, LogEntry{
			Level:   ErrorLevel,
			Tag:     "ALERTER_ERROR",
			Message: fmt.Sprintf("Failed to send alarm via alerter: %v", sendErr),
			Fields: map[string]interface{}{
				"origin_level": level,
				"origin_tag":   tag,
				"origin_msg":   msg,
				"origin_err":   err,
			},
		})
	}
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

	if studentID := c.Value(constant.CtxStudentID); studentID != nil {
		content = append(content, fmt.Sprintf("【STUDENT ID】: %v", studentID))
	}

	if teacherID := c.Value(constant.CtxTeacherID); teacherID != nil {
		content = append(content, fmt.Sprintf("【TEACHER ID】: %v", teacherID))
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
