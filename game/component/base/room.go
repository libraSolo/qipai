package base

import "game/component/proto"

type RoomFrame interface {
	GetUsers() map[string]*proto.RoomUser
	GetID() string
}
