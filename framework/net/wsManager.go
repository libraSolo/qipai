package net

import (
	"common/logs"
	"common/utils"
	"encoding/json"
	"errors"
	"fmt"
	"framework/game"
	"framework/protocol"
	"framework/remote"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/rand"
	"net/http"
	"strings"
	"sync"
	"time"
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
	RemoteReadChan     chan []byte
	RemoteCli          remote.Client
	RemotePushChan     chan *remote.Msg
}

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]*WsConnection),
		handles:        make(map[protocol.PackageType]EventHandler),
		RemoteReadChan: make(chan []byte, 1024),
		RemotePushChan: make(chan *remote.Msg, 1024),
	}
}

func (m *Manager) Run(addr string) {
	go m.clientReadChanHandler()
	go m.remoteReadChanHandler()
	go m.remotePushChanHandler()
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

func (m *Manager) clientReadChanHandler() {
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

func (m *Manager) remoteReadChanHandler() {
	for {
		select {
		case body, ok := <-m.RemoteReadChan:
			//logs.Info("sub nats msg received:%v", string(msg))
			if ok {
				m.decodeNatsPack(body)
			}
		}
	}
}

func (m *Manager) decodeNatsPack(body []byte) {
	var msg remote.Msg
	err := json.Unmarshal(body, &msg)
	if err != nil {
		logs.Error("decode nats msg err: %v", err)
		return
	}
	if msg.Type == remote.SessionType {
		// 特殊处理 session类型存在 connection中,不推送到客户端
	}
	if msg.Body != nil {
		if msg.Body.Type == protocol.Request {
			// 给客户端回信息: res
			msg.Body.Type = protocol.Response
			m.Response(&msg)
		}
		if msg.Body.Type == protocol.Push {
			m.RemotePushChan <- &msg
		}
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
		dst, err := m.selectDst(serverType)
		if err != nil {
			logs.Error("remote select dst err")
			return err
		}

		msg := remote.Msg{
			Cid:         c.GetSession().ClientId,
			Uid:         c.GetSession().Uid,
			Src:         m.ServerId,
			Dst:         dst,
			Router:      handlerMethod,
			Body:        message,
			SessionData: c.GetSession().data,
		}
		data, _ := json.Marshal(msg)
		err = m.RemoteCli.SendMsg(dst, data)
		if err != nil {
			logs.Error("remote send msg err %v", err)
			return err
		}
	}
	return nil
}

func (m *Manager) HandKickHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("kick...")

	return nil
}

func (m *Manager) selectDst(serverType string) (string, error) {
	configs, ok := game.Conf.ServersConf.TypeServer[serverType]
	if !ok {
		return "", errors.New("no server")
	}
	// 负载均衡 || 随机
	rand.Seed(uint64(time.Now().Unix()))
	index := rand.Intn(len(configs))
	return configs[index].ID, nil
}

func (m *Manager) remotePushChanHandler() {
	for {
		select {
		case body, ok := <-m.RemotePushChan:
			if ok {
				if body.Body.Type == protocol.Push {
					m.Response(body)
				}
			}
		}
	}
}

func (m *Manager) Response(msg *remote.Msg) {
	connection, ok := m.clients[msg.Cid]
	if !ok {
		logs.Info("%s client down, uid=%s", msg.Cid)
		return
	}
	encode, err := protocol.MessageEncode(msg.Body)
	if err != nil {
		logs.Error("res message encode error:%v", err)
		return
	}
	res, err := protocol.Encode(protocol.Data, encode)
	if err != nil {
		logs.Error("res message encode error:%v", err)
		return
	}
	if msg.Body.Type == protocol.Push {
		for _, connection = range m.clients {
			if utils.Contains(msg.PushUser, connection.GetSession().Uid) {
				err := connection.SendMessage(res)
				if err != nil {
					logs.Error("res message send error:%v", err)
					continue
				}
			}
		}
		return
	}
	err = connection.SendMessage(res)
	if err != nil {
		logs.Error("res message send error:%v", err)
	}
}
