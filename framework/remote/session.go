package remote

type Session struct {
	client Client
	msg    *Msg
}

func NewSession(cli Client, msg *Msg) *Session {
	return &Session{
		client: cli,
		msg:    msg,
	}
}

func (s *Session) GetUid() string {
	return s.msg.Uid
}
