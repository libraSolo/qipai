package room

import (
	"framework/remote"
	"game/component/proto"
)

type GameFrame interface {
	GetGameData() any
	StartGame(session *remote.Session, user *proto.RoomUser)
}
