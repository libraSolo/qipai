package node

import (
	"common/logs"
	"encoding/json"
	"framework/remote"
)

// App nats 的客户端 处理实际游戏的逻辑
type App struct {
	remoteCli remote.Client
	readChan  chan []byte
	writeChan chan *remote.Msg
	Handlers  LogicHandler
}

func Default() *App {
	return &App{
		readChan:  make(chan []byte),
		writeChan: make(chan *remote.Msg),
		Handlers:  make(LogicHandler),
	}
}

func (a *App) Run(serverId string) error {
	a.remoteCli = remote.NewNatsClient(serverId, a.readChan)
	err := a.remoteCli.Run()
	if err != nil {
		return err
	}
	go a.readChanMsg()
	go a.writeChanMsg()
	return nil
}

func (a *App) readChanMsg() {
	// 收到其他nats 发的消息
	for {
		select {
		case msg := <-a.readChan:
			var remoteMsg remote.Msg
			err := json.Unmarshal(msg, &remoteMsg)
			if err != nil {
				logs.Error("Error unmarsh remote message err:%v ", err)
				continue
			}
			session := remote.NewSession(a.remoteCli, &remoteMsg)
			session.SetData(remoteMsg.SessionData)
			// 路由分发
			router := remoteMsg.Router
			if handlerFunc := a.Handlers[router]; handlerFunc != nil {
				result := handlerFunc(session, remoteMsg.Body.Data)
				if result != nil {
					remoteMsg.Body.Data, _ = json.Marshal(result)
				}
				responseMsg := &remote.Msg{
					Cid:  remoteMsg.Cid,
					Body: remoteMsg.Body,
					Src:  remoteMsg.Dst,
					Dst:  remoteMsg.Src,
					Uid:  remoteMsg.Uid,
				}
				a.writeChan <- responseMsg
			}
		}
	}
}

func (a *App) writeChanMsg() {
	for {
		select {
		case msg, ok := <-a.writeChan:
			if ok {
				marshal, _ := json.Marshal(msg)
				err := a.remoteCli.SendMsg(msg.Dst, marshal)
				if err != nil {
					logs.Error("app remote send msg error: ", err)
				}
			}

		}
	}
}

func (a *App) Close() {
	if a.remoteCli != nil {
		a.remoteCli.Close()
	}
}

func (a *App) RegisterHandler(register LogicHandler) {
	a.Handlers = register
}
