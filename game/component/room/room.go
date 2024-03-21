package room

import (
	"core/models/entity"
	"framework/remote"
)

type Room struct {
	Id string
}

func NewRoom(id string) *Room {
	return &Room{
		Id: id,
	}
}

func (r *Room) UserEntryRoom(session *remote.Session, user *entity.User) error {
	// 房间号  推送给客户端
	r.UpdateUserInfoRoomPush(session, user.Uid)
	// 游戏类型推送給客戶端
	return nil
}

func (r *Room) UpdateUserInfoRoomPush(session *remote.Session, uid string) {
	pushMsg := map[string]any{
		"roomID":     r.Id,
		"pushRouter": "UpdateUserInfoPush",
	}
	// node 是 nats client -> 通过 nats 将消息发给 connector 服务 -> 发给客户端
	session.Push([]string{uid}, pushMsg, "ServerMessagePush")
}
