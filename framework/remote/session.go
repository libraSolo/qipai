package remote

import (
	"common/logs"
	"encoding/json"
	"framework/protocol"
)

type Session struct {
	client   Client
	msg      *Msg
	pushChan chan *userPushMsg
}

type pushMsg struct {
	data   []byte
	router string
}

type userPushMsg struct {
	PushMsg pushMsg  `json:"pushMsg"`
	User    []string `json:"user"`
}

func NewSession(cli Client, msg *Msg) *Session {
	s := &Session{
		client:   cli,
		msg:      msg,
		pushChan: make(chan *userPushMsg, 1024),
	}
	go s.pushChanRead()
	return s
}

func (s *Session) GetUid() string {
	return s.msg.Uid
}

func (s *Session) Push(user []string, data any, router string) {
	marshal, _ := json.Marshal(data)
	uMsg := &userPushMsg{
		PushMsg: pushMsg{
			data:   marshal,
			router: router,
		},
		User: user,
	}
	s.pushChan <- uMsg
}

func (s *Session) pushChanRead() {
	for {
		select {
		case data := <-s.pushChan:
			pushMsg := &protocol.Message{
				Type:  protocol.Push,
				ID:    s.msg.Body.ID,
				Route: data.PushMsg.router,
				Data:  data.PushMsg.data,
			}
			msg := Msg{
				Dst:      s.msg.Src,
				Src:      s.msg.Dst,
				Body:     pushMsg,
				Cid:      s.msg.Cid,
				Uid:      s.msg.Uid,
				PushUser: data.User,
			}
			res, _ := json.Marshal(msg)
			err := s.client.SendMsg(msg.Dst, res)
			if err != nil {
				logs.Error("push message failed err: %v ,message:%v", err, msg)
			}
		}
	}
}
