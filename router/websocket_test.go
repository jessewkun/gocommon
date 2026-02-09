package router

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type testWSHandler struct {
	onConnectErr   error
	onMessageCount int32
	onPingCount    int32
	onDisconnect   int32
}

func (h *testWSHandler) OnConnect(c *gin.Context, conn *websocket.Conn) error {
	return h.onConnectErr
}

func (h *testWSHandler) OnMessage(c *gin.Context, messageType int, message []byte) {
	atomic.AddInt32(&h.onMessageCount, 1)
}

func (h *testWSHandler) OnDisconnect(c *gin.Context, conn *websocket.Conn) {
	atomic.AddInt32(&h.onDisconnect, 1)
}

func (h *testWSHandler) OnPing(c *gin.Context, conn *websocket.Conn) {
	atomic.AddInt32(&h.onPingCount, 1)
}

func newWSServer(t *testing.T, handler WebSocketHandler, cfg *WsConfig) (string, string, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		WsWrapperWithConfig(c, handler, cfg)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping websocket tests; listen not permitted: %v", err)
	}
	srv := httptest.NewUnstartedServer(r)
	srv.Listener = ln
	srv.Start()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, srv.URL, srv.Close
}

func dialWS(t *testing.T, wsURL, origin string) *websocket.Conn {
	t.Helper()
	header := http.Header{}
	header.Set("Origin", origin)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	return conn
}

func TestWsWrapperWithConfig_NilConfig_PingPong(t *testing.T) {
	handler := &testWSHandler{}
	wsURL, origin, closeFn := newWSServer(t, handler, nil)
	defer closeFn()

	conn := dialWS(t, wsURL, origin)
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("write ping failed: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	mt, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read pong failed: %v", err)
	}
	if mt != websocket.TextMessage || string(msg) != "pong" {
		t.Fatalf("unexpected response: type=%d msg=%s", mt, string(msg))
	}

	if got := atomic.LoadInt32(&handler.onPingCount); got != 1 {
		t.Fatalf("expected OnPing called once, got %d", got)
	}
}

func TestWsWrapperWithConfig_OnDisconnect_Once(t *testing.T) {
	handler := &testWSHandler{}
	wsURL, origin, closeFn := newWSServer(t, handler, nil)
	defer closeFn()

	conn := dialWS(t, wsURL, origin)
	_ = conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&handler.onDisconnect) == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected OnDisconnect called once, got %d", atomic.LoadInt32(&handler.onDisconnect))
}

func TestWsWrapperWithConfig_OnConnectError(t *testing.T) {
	handler := &testWSHandler{onConnectErr: errors.New("connect failed")}
	wsURL, origin, closeFn := newWSServer(t, handler, nil)
	defer closeFn()

	conn := dialWS(t, wsURL, origin)
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatalf("expected read error after OnConnect failure")
	}
	if got := atomic.LoadInt32(&handler.onDisconnect); got != 0 {
		t.Fatalf("expected OnDisconnect not called, got %d", got)
	}
}
