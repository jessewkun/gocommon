// Package router 提供 WebSocket 支持
package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"
)

// WebSocketHandler 定义 WebSocket 处理器接口
// messageType 为 websocket.TextMessage 或 websocket.BinaryMessage
type WebSocketHandler interface {
	OnConnect(c *gin.Context, conn *websocket.Conn) error
	OnMessage(c *gin.Context, messageType int, message []byte)
	OnDisconnect(c *gin.Context, conn *websocket.Conn)
	OnPing(c *gin.Context, conn *websocket.Conn)
}

const (
	defaultReadTimeout = 90 * time.Second
)

// WsConfig WebSocket 服务端配置，可选
type WsConfig struct {
	// CheckOrigin 校验请求 Origin，用于跨域。为 nil 时使用 gorilla/websocket 默认（Origin 与 Host 一致或未设置 Origin 才通过）
	CheckOrigin func(r *http.Request) bool
	// ReadTimeout 读超时，超时未收到消息将断开。为零值时使用 defaultReadTimeout
	ReadTimeout time.Duration
}

// readTimeout 读超时
func (c *WsConfig) getReadTimeout() time.Duration {
	if c != nil && c.ReadTimeout > 0 {
		return c.ReadTimeout
	}
	return defaultReadTimeout
}

// checkOrigin 校验请求 Origin，用于跨域
func (c *WsConfig) checkOrigin() func(r *http.Request) bool {
	if c != nil && c.CheckOrigin != nil {
		return c.CheckOrigin
	}
	return nil
}

// CheckOriginFromAllowList 根据允许的 Origin 列表生成 CheckOrigin 函数，供 WsConfig 使用
// 若 origins 为空则拒绝所有带 Origin 的请求；无 Origin 头的请求（如直接访问）放行
func CheckOriginFromAllowList(origins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		for _, allowed := range origins {
			if allowed == origin {
				return true
			}
		}
		return false
	}
}

// upgrader 根据配置生成，避免包级变量依赖业务 config
func newUpgrader(cfg *WsConfig) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: cfg.checkOrigin(),
	}
}

// WsWrapper 使用默认配置处理 WebSocket 请求（CheckOrigin 为 nil，ReadTimeout 90s）
func WsWrapper(c *gin.Context, handler WebSocketHandler) {
	WsWrapperWithConfig(c, handler, nil)
}

// WsWrapperWithConfig 使用指定配置处理 WebSocket 请求，支持跨域与读超时配置
func WsWrapperWithConfig(c *gin.Context, handler WebSocketHandler, cfg *WsConfig) {
	cfg = normalizeWsConfig(cfg)
	timeout := cfg.getReadTimeout()

	conn, ok := upgradeWs(c, cfg)
	if !ok {
		return
	}
	defer conn.Close()

	if !handleWsConnect(c, handler, conn) {
		return
	}

	state := newWsState()
	writeControl, writeMessage := newWsWriters(conn, &state.writeMutex)
	setupWsHandlers(c, handler, conn, timeout, writeControl, state)
	runWsReadLoop(c, handler, conn, writeControl, writeMessage, state)
}

// normalizeWsConfig 规范化配置，如果配置为 nil，则返回默认配置
func normalizeWsConfig(cfg *WsConfig) *WsConfig {
	if cfg == nil {
		return &WsConfig{}
	}
	return cfg
}

// upgradeWs 升级 WebSocket 连接
func upgradeWs(c *gin.Context, cfg *WsConfig) (*websocket.Conn, bool) {
	upgrader := newUpgrader(cfg)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.SystemError(c)
		return nil, false
	}
	return conn, true
}

// handleWsConnect 处理 WebSocket 连接建立
func handleWsConnect(c *gin.Context, handler WebSocketHandler, conn *websocket.Conn) bool {
	if err := handler.OnConnect(c, conn); err != nil {
		logger.Error(c.Request.Context(), "WsWrapper", err)
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()),
			time.Now().Add(time.Second))
		_ = conn.Close()
		return false
	}
	return true
}

// wsState WebSocket 状态
type wsState struct {
	connectionClosed bool
	closeMutex       sync.Mutex
	writeMutex       sync.Mutex
	disconnectOnce   sync.Once
}

// newWsState 创建 WebSocket 状态
func newWsState() *wsState {
	return &wsState{}
}

// newWsWriters 创建 WebSocket 写入器
func newWsWriters(conn *websocket.Conn, writeMutex *sync.Mutex) (
	func(messageType int, data []byte) error,
	func(messageType int, data []byte) error,
) {
	writeControl := func(messageType int, data []byte) error {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		return conn.WriteControl(messageType, data, time.Now().Add(time.Second))
	}
	writeMessage := func(messageType int, data []byte) error {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		return conn.WriteMessage(messageType, data)
	}
	return writeControl, writeMessage
}

// setupWsHandlers 设置 WebSocket 处理器
func setupWsHandlers(
	c *gin.Context,
	handler WebSocketHandler,
	conn *websocket.Conn,
	timeout time.Duration,
	writeControl func(messageType int, data []byte) error,
	state *wsState,
) {
	conn.SetCloseHandler(func(code int, text string) error {
		state.closeMutex.Lock()
		state.connectionClosed = true
		state.closeMutex.Unlock()
		logger.InfoWithField(c.Request.Context(), "WsWrapper", "WebSocket connection closed", map[string]interface{}{
			"code": code,
			"text": text,
		})
		state.disconnectOnce.Do(func() { handler.OnDisconnect(c, conn) })
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(timeout))
		return writeControl(websocket.PongMessage, []byte(appData))
	})

	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(timeout))
		return nil
	})

	conn.SetReadDeadline(time.Now().Add(timeout))
}

// runWsReadLoop 运行 WebSocket 读取循环
func runWsReadLoop(
	c *gin.Context,
	handler WebSocketHandler,
	conn *websocket.Conn,
	writeControl func(messageType int, data []byte) error,
	writeMessage func(messageType int, data []byte) error,
	state *wsState,
) {
	for {
		state.closeMutex.Lock()
		if state.connectionClosed {
			state.closeMutex.Unlock()
			break
		}
		state.closeMutex.Unlock()

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			state.closeMutex.Lock()
			if state.connectionClosed {
				state.closeMutex.Unlock()
				break
			}
			state.closeMutex.Unlock()

			logger.InfoWithField(c.Request.Context(), "WsWrapper", "WebSocket abnormal closure", map[string]interface{}{
				"error": err.Error(),
			})
			_ = writeControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "连接异常断开"))
			state.disconnectOnce.Do(func() { handler.OnDisconnect(c, conn) })
			break
		}

		if messageType == websocket.TextMessage {
			if string(message) == "ping" {
				if err := writeMessage(websocket.TextMessage, []byte("pong")); err != nil {
					logger.Error(c.Request.Context(), "WsWrapper", err)
				}
				handler.OnPing(c, conn)
				continue
			}
		}
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			handler.OnMessage(c, messageType, message)
		} else {
			logger.InfoWithField(c.Request.Context(), "WsWrapper", "received non-text/binary message", map[string]interface{}{
				"messageType": messageType,
				"messageSize": len(message),
			})
		}
	}
}
