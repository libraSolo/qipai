package base

import (
	"framework/remote"
	"game/component/proto"
)

type RoomFrame interface {
	GetUsers() map[string]*proto.RoomUser
	GetID() string
	EndGame(session *remote.Session)
	UserReady(session *remote.Session, uid string)
}
