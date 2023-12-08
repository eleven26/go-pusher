package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"go-websocket/message"
)

// bufferSize 通道缓冲区、map 初始化大小
const bufferSize = 128

// Handler 错误处理函数
type Handler func(log message.Log, err error)

// Hub 维护了所有的客户端连接
type Hub struct {
	// 注册请求
	register chan *Client

	// 取消注册请求
	unregister chan *Client

	// 记录 uid 跟 client 的对应关系
	userClients map[string]*Client

	// 互斥锁，保护 userClients 以及 clients 的读写
	sync.RWMutex

	// 消息记录器
	messageLogger message.Logger

	// 错误处理器
	errorHandler Handler

	// 验证器
	authenticator Authenticator

	// 等待发送的消息数量
	pending atomic.Int64
}

// 默认的错误处理器
func defaultErrorHandler(log message.Log, err error) {
	res, _ := json.Marshal(log)
	fmt.Printf("send message: %s, error: %s\n", string(res), err.Error())
}

func newHub() *Hub {
	return &Hub{
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		userClients:   make(map[string]*Client, bufferSize),
		RWMutex:       sync.RWMutex{},
		messageLogger: &message.StdoutMessageLogger{},
		errorHandler:  defaultErrorHandler,
		authenticator: &JWTAuthenticator{},
	}
}

// 注册、取消注册请求处理
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.userClients[client.uid] = client
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			close(client.send)
			delete(h.userClients, client.uid)
			h.Unlock()
		}
	}
}

// 返回 Hub 的当前的关键指标
func metrics(hub *Hub, w http.ResponseWriter) {
	pending := hub.pending.Load()
	connections := len(hub.userClients)
	_, _ = w.Write([]byte(fmt.Sprintf("# HELP connections 连接数\n# TYPE connections gauge\nconnections %d\n", connections)))
	_, _ = w.Write([]byte(fmt.Sprintf("# HELP pending 等待发送的消息数量\n# TYPE pending gauge\npending %d\n", pending)))
}
