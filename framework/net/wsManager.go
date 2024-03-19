package net

import (
	"common/logs"
	"encoding/json"
	"errors"
	"fmt"
	"framework/game"
	"framework/protocol"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
)

var (
	webSocketUpgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type CheckOriginHandler func(r *http.Request) bool // 跨域
type EventHandler func(packet *protocol.Packet, c Connection) error
type HandlerFunc func(session *Session, body []byte) (any, error)
type LogicHandler map[string]HandlerFunc // 逻辑处理器

type Manager struct {
	sync.RWMutex
	ServerId           string
	webSocketUpgrade   *websocket.Upgrader
	CheckOriginHandler CheckOriginHandler
	clients            map[string]*WsConnection
	ClientReadChan     chan *MsgPack
	handles            map[protocol.PackageType]EventHandler
	ConnectorHandlers  LogicHandler
}

func (m *Manager) Run(addr string) {
	go m.ClientReadChanHandler()
	http.HandleFunc("/", m.serverWS)

	// 设置不同的消息处理器
	m.setupEventHandlers()
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logs.Fatal("connect to server err:%v", err)
	}
}

func (m *Manager) serverWS(writer http.ResponseWriter, request *http.Request) {
	// websocket 基于 http
	if m.webSocketUpgrade == nil {
		m.webSocketUpgrade = &webSocketUpgrade
	}
	wsConn, err := m.webSocketUpgrade.Upgrade(writer, request, nil)
	if err != nil {
		logs.Error("websocket upgrade err:%v", err)
		return
	}
	client := NewWsConnection(wsConn, m)
	m.addClient(client)
	client.Run()
}

func (m *Manager) addClient(client *WsConnection) {
	m.Lock()
	defer m.Unlock()

	m.clients[client.ClientId] = client
}

func (m *Manager) removeClient(wsC *WsConnection) {
	for cid, c := range m.clients {
		if cid == wsC.ClientId {
			c.Close()
			delete(m.clients, cid)
		}
	}
}

func (m *Manager) ClientReadChanHandler() {
	for {
		select {
		case body, ok := <-m.ClientReadChan:
			if ok {
				m.decodeClientPack(body)
			}
		}
	}
}

func (m *Manager) decodeClientPack(body *MsgPack) {
	// 解析协议
	packet, err := protocol.Decode(body.Body)
	if err != nil {
		logs.Error("decode message err: %v", err)
		return
	}
	if err := m.routeEvent(packet, body.ClientId); err != nil {
		logs.Error("no route found err: %v", err)
		return
	}

}

func (m *Manager) Close() {
	for _, v := range m.clients {
		v.Close()
		delete(m.clients, v.ClientId)
	}
}

func (m *Manager) routeEvent(packet *protocol.Packet, clientId string) error {
	// 根据 type 处理
	conn, ok := m.clients[clientId]
	if ok {
		handler, ok := m.handles[packet.Type]
		if ok {
			return handler(packet, conn)
		}
		return errors.New("no packetType found for client")
	}
	return errors.New("no client found")
}

func (m *Manager) setupEventHandlers() {
	m.handles[protocol.Handshake] = m.HandshakeHandler
	m.handles[protocol.HandshakeAck] = m.HandshakeAckHandler
	m.handles[protocol.Heartbeat] = m.HeartbeatHandler
	m.handles[protocol.Data] = m.MessageHandler
	m.handles[protocol.Kick] = m.HandKickHandler
}

func (m *Manager) HandshakeHandler(packet *protocol.Packet, c Connection) error {
	res := protocol.HandshakeResponse{
		Code: 200,
		Sys: protocol.Sys{
			Heartbeat: 3,
		},
	}
	data, _ := json.Marshal(res)
	buf, err := protocol.Encode(packet.Type, data)
	if err != nil {
		logs.Error("encode packet err:%v", err)
		return err
	}
	return c.SendMessage(buf)
}

func (m *Manager) HandshakeAckHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("HandshakeAck...")
	return nil
}

func (m *Manager) HeartbeatHandler(packet *protocol.Packet, c Connection) error {
	var res []byte
	data, _ := json.Marshal(res)
	buf, err := protocol.Encode(packet.Type, data)
	if err != nil {
		logs.Error("encode packet err:%v", err)
		return err
	}
	return c.SendMessage(buf)
}

func (m *Manager) MessageHandler(packet *protocol.Packet, c Connection) error {
	message := packet.MessageBody()
	// connector.entryHandler.entry
	routeStr := message.Route
	routers := strings.Split(routeStr, ".")
	if 3 != len(routers) {
		return errors.New("router unsupported")
	}
	serverType := routers[0]
	handlerMethod := fmt.Sprintf("%s.%s", routers[1], routers[2])
	connectorConfig := game.Conf.GetConnectorByServerType(serverType)
	if nil != connectorConfig {
		// 本地 connector 服务器
		handlerFunc, ok := m.ConnectorHandlers[handlerMethod]
		if ok {
			data, err := handlerFunc(c.GetSession(), message.Data)
			if err != nil {
				return err
			}
			marshal, _ := json.Marshal(data)
			message.Type = protocol.Response
			message.Data = marshal
			encode, err := protocol.MessageEncode(message)
			if err != nil {
				return err
			}
			res, err := protocol.Encode(packet.Type, encode)
			if err != nil {
				return err
			}
			return c.SendMessage(res)
		}
	} else {
		// nats 远端调用
	}
	return nil
}

func (m *Manager) HandKickHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("kick...")

	return nil
}

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]*WsConnection),
		handles:        make(map[protocol.PackageType]EventHandler),
	}
}
