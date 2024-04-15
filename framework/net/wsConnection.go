package net

import (
	"common/logs"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"time"
)

var (
	baseClientId   uint64 = 10000
	maxMessageSize int64  = 1024
	pongWait              = 10 * time.Second
	writeWait             = 10 * time.Second
	pingInterval          = pongWait * 8 / 10
)

type WsConnection struct {
	ClientId   string
	Conn       *websocket.Conn
	manager    *Manager
	ReadChan   chan *MsgPack
	WriteChan  chan []byte
	Session    *Session
	pingTicker *time.Ticker
}

func NewWsConnection(conn *websocket.Conn, manager *Manager) *WsConnection {
	cid := fmt.Sprintf("%s-%s-%d", uuid.New().String(), manager.ServerId, atomic.AddUint64(&baseClientId, 1))
	return &WsConnection{
		ClientId:  cid,
		Conn:      conn,
		manager:   manager,
		WriteChan: make(chan []byte, 1024),
		ReadChan:  manager.ClientReadChan,
		Session:   NewSession(cid),
	}
}

func (c *WsConnection) Run() {
	go c.readMessage()
	go c.writeMessage()
	// 心跳检测 ws: ping/pong
	c.Conn.SetPongHandler(c.PongHandler)
}

func (c *WsConnection) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logs.Error("readMessage dead line err: %v", err)
		return
	}
	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// 客戶端发的消息是二进制
		if messageType == websocket.BinaryMessage {
			if c.ReadChan != nil {
				c.ReadChan <- &MsgPack{
					ClientId: c.ClientId,
					Body:     message,
				}
			}
		} else {
			// 其他消息
			logs.Error("unknown message type : %d", messageType)
		}
	}
}

func (c *WsConnection) writeMessage() {

	c.pingTicker = time.NewTicker(pingInterval)
	for {
		select {
		case <-c.pingTicker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logs.Error("client[%s] ping dead line error :%v", c.ClientId, err)

			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logs.Error("client[%s] ping error :%v", c.ClientId, err)
				c.Close()
			}
		case message, ok := <-c.WriteChan:
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					logs.Error("connection closed, %v", err)
				}
				return
			}
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				logs.Info("client[%s] write message error :%v", c.ClientId, err)
			}
		}

	}
}

func (c *WsConnection) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
}

func (c *WsConnection) SendMessage(buf []byte) error {
	c.WriteChan <- buf
	return nil
}

func (c *WsConnection) GetSession() *Session {
	return c.Session
}

func (c *WsConnection) PongHandler(data string) error {
	//logs.Info("pong...")
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logs.Error("pong err")
		return err
	}
	return nil
}
