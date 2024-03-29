package remote

import (
	"common/logs"
	"encoding/json"
	"framework/protocol"
	"sync"
)

type Session struct {
	sync.RWMutex
	client          Client
	msg             *Msg
	pushChan        chan *userPushMsg
	data            map[string]any
	pushSessionChan chan map[string]any
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
		client:          cli,
		msg:             msg,
		pushChan:        make(chan *userPushMsg, 1024),
		data:            make(map[string]any),
		pushSessionChan: make(chan map[string]any, 1024),
	}
	go s.pushChanRead()
	go s.pushSessionChanRead()
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

func (s *Session) Put(key string, value string) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = value
	s.pushSessionChan <- s.data
}

func (s *Session) pushSessionChanRead() {
	for {
		select {
		case data := <-s.pushSessionChan:
			msg := Msg{
				Dst:         s.msg.Src,
				Src:         s.msg.Dst,
				Cid:         s.msg.Cid,
				Uid:         s.msg.Uid,
				SessionData: data,
				Type:        SessionType,
			}
			res, _ := json.Marshal(msg)
			if err := s.client.SendMsg(msg.Dst, res); err != nil {
				logs.Error("push session data err:%v", err)
			}
		}
	}
}

func (s *Session) SetData(data map[string]any) {
	s.Lock()
	defer s.Unlock()
	for k, v := range data {
		s.data[k] = v
	}
}

func (s *Session) Get(k string) (any, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.data[k]
	return v, ok
}
