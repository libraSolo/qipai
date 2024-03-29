package net

import "sync"

type Session struct {
	sync.RWMutex
	ClientId string
	Uid      string
	data     map[string]any
}

func NewSession(cid string) *Session {
	return &Session{
		ClientId: cid,
		data:     make(map[string]any),
	}
}

func (s *Session) Put(key string, value any) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = value
}

func (s *Session) Get(key string) (value any, ok bool) {
	s.RLock()
	defer s.RUnlock()
	value, ok = s.data[key]
	return
}

func (s *Session) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func (s *Session) SetData(uid string, data map[string]any) {
	s.Lock()
	defer s.Unlock()
	if s.Uid == uid {
		for k, v := range data {
			s.data[k] = v
		}
	}
}
