package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	_ = os.Setenv("JWT_SECRET", _testSecret)

	hub := newHub()
	go hub.run()

	server := httptest.NewServer(mux(hub))
	defer server.Close()

	// 等待一段时间确保服务器已经启动
	time.Sleep(100 * time.Millisecond)

	// 创建一个 WebSocket 连接
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws?token="+_testToken, nil)
	if err != nil {
		t.Fatalf("Error dialing WebSocket: %v", err)
	}
	defer conn.Close()

	// 等待连接成功
	time.Sleep(100 * time.Millisecond)

	// /ws
	assert.Equal(t, 0, len(hub.register))
	assert.Equal(t, 0, len(hub.unregister))
	assert.Equal(t, 1, len(hub.userClients))
	client, ok := hub.userClients[_testUid]
	assert.True(t, ok)

	// /send
	sendHelloWorld(server.URL)
	assert.Equal(t, 0, len(client.send))
	assert.Equal(t, int64(0), hub.pending.Load())

	// 客户端应该收到消息
	_, msg, err := conn.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(msg))

	// /metrics
	assert.Equal(t, 1, len(hub.userClients))
	resp, _ := http.Get(server.URL + "/metrics")
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "# HELP connections 连接数\n# TYPE connections gauge\nconnections 1\n# HELP pending 等待发送的消息数量\n# TYPE pending gauge\npending 0\n", string(body))

	server.Close()
}

func sendHelloWorld(url string) {
	url = url + "/send?uid=123&message=Hello%20World"
	method := "GET"

	req, _ := http.NewRequest(method, url, nil)

	client := &http.Client{}
	_, _ = client.Do(req)
}
