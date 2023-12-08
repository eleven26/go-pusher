package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-websocket/message"

	"github.com/gorilla/websocket"
)

const (
	// 推送消息的超时时间
	writeWait = 10 * time.Second

	// 允许客户端发送的最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 我们可以在这里检查请求来源是否合法
		// 这里我们直接返回 true，表示接受所有请求
		return true
	},
}

// Client 是一个中间人，负责 WebSocket 连接和 Hub 之间的通信
type Client struct {
	hub *Hub

	// 底层的 WebSocket 连接
	conn *websocket.Conn

	// 缓冲发送消息的通道
	send chan message.Log

	// 关联的用户 id
	uid string
}

// 连接关闭时的处理函数
// 正常的断开不做处理，非正常的断开打印日志
func closeHandler(code int, text string) error {
	if code >= 1002 {
		log.Println("connection close: ", code, text)
	}
	return nil
}

// readPump 从 WebSocket 连接中读取消息
//
// 该方法在一个独立的协程中运行，我们保证了每个连接只有一个 reader。
// 该方法会丢弃所有客户端传来的消息，如果需要接收可以在这里进行处理。
func (c *Client) readPump() {
	defer func() {
		// unregister 为无缓冲通道，下面这一行会阻塞，
		// 直到 hub.run 中的 <-h.unregister 语句执行
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Time{}) // 永不超时
	for {
		// 从客户端接收消息
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump 负责推送消息给 WebSocket 客户端
//
// 该方法在一个独立的协程中运行，我们保证了每个连接只有一个 writer。
// Client 会从 send 请求中获取消息，然后在这个方法中推送给客户端。
func (c *Client) writePump() {
	defer func() {
		_ = c.conn.Close()
	}()

	// 从 send 通道中获取消息，然后推送给客户端
	for {
		messageLog, ok := <-c.send

		// 设置写超时时间
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		// c.send 这个通道已经被关闭了
		if !ok {
			c.hub.pending.Add(int64(-1 * len(c.send)))
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, StringToBytes(messageLog.Message)); err != nil {
			c.hub.errorHandler(messageLog, err)
			c.hub.pending.Add(int64(-1 * len(c.send)))
			return
		}

		c.hub.pending.Add(int64(-1))
	}
}

// serveWs 处理 WebSocket 连接请求
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("upgrade error: %s", err.Error())))
		return
	}

	// 认证失败的时候，返回错误信息，并断开连接
	uid, err := hub.authenticator.Authenticate(r)
	if err != nil {
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("authenticate error: %s", err.Error())))
		_ = conn.Close()
		return
	}

	// 注册 Client
	client := &Client{hub: hub, conn: conn, send: make(chan message.Log, bufferSize), uid: uid}
	client.conn.SetCloseHandler(closeHandler)
	// register 无缓冲，下面这一行会阻塞，直到 hub.run 中的 <-h.register 语句执行
	// 这样可以保证 register 成功之后才会启动读写协程
	client.hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// send 处理消息发送请求
func send(hub *Hub, w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	if uid == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("uid is required"))
		return
	}

	// 从 hub 中获取 uid 关联的 client
	hub.RLock()
	client, ok := hub.userClients[uid]
	hub.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("client not found: %s", uid)))
		return
	}

	// 记录消息
	messageLog := message.Log{Uid: uid, Message: r.FormValue("message")}
	_ = hub.messageLogger.Log(messageLog)

	// 发送消息
	client.send <- messageLog

	// 增加等待发送的消息数量
	hub.pending.Add(int64(1))
}
