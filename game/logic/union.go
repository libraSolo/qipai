package logic

import (
	"core/models/entity"
	"core/service"
	"framework/errorCode"
	"framework/remote"
	"game/component/room"
	"game/models/request"
	"sync"
)

type Union struct {
	sync.RWMutex
	Id       int64
	manager  *UnionManager
	RoomList map[string]*room.Room
}

func NewUnion(m *UnionManager) *Union {
	return &Union{
		manager:  m,
		RoomList: make(map[string]*room.Room),
	}
}

func (u *Union) CreateRoom(service *service.UserService, session *remote.Session, req request.CreateRoomReq, user *entity.User) *errorCode.Error {
	// 创建一个房间 生成一个房间号
	id := u.manager.CreateRoomId()
	newRoom := room.NewRoom(id, req.UnionID, req.GameRule, u)
	u.RoomList[id] = newRoom

	return newRoom.UserEntryRoom(session, user)
}

func (u *Union) DismissRoom(roomId string) {
	u.Lock()
	defer u.Unlock()
	delete(u.RoomList, roomId)
}
