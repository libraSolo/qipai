package room

import (
	"core/models/entity"
	"framework/remote"
	"game/component/proto"
	"game/component/sz"
	"game/models/request"
)

type Room struct {
	Id          string
	unionID     int64
	gameRule    proto.GameRule
	users       map[string]*proto.RoomUser
	RoomCreator *proto.RoomCreator
	GameFrame   GameFrame
}

func NewRoom(id string, unionID int64, rule proto.GameRule) *Room {
	room := &Room{
		Id:       id,
		unionID:  unionID,
		gameRule: rule,
		users:    make(map[string]*proto.RoomUser),
	}
	if rule.GameType == int(proto.PinSanZhang) {
		room.GameFrame = sz.NewGameFrame(room, rule)
	}
	return room
}

func (r *Room) UserEntryRoom(session *remote.Session, user *entity.User) error {
	r.RoomCreator = &proto.RoomCreator{
		Uid: user.Uid,
	}
	if r.unionID == 1 {
		r.RoomCreator.CreatorType = proto.UserCreatorType
	} else {
		r.RoomCreator.CreatorType = proto.UnionCreatorType
	}
	r.users[user.Uid] = &proto.RoomUser{
		UserInfo:   *proto.ToRoomUser(user),
		ChairId:    0,
		UserStatus: proto.None,
	}
	// 房间号  推送给客户端
	r.UpdateUserInfoRoomPush(session, user.Uid)
	session.Put("roomId", r.Id)
	// 游戏类型推送給客戶端
	r.SelfEntryRoomPush(session, user.Uid)
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

func (r *Room) SelfEntryRoomPush(session *remote.Session, uid string) {
	pushMsg := map[string]any{
		"gameType":   r.gameRule.GameType,
		"pushRouter": "SelfEntryRoomPush",
	}
	// node 是 nats client -> 通过 nats 将消息发给 connector 服务 -> 发给客户端
	session.Push([]string{uid}, pushMsg, "ServerMessagePush")
}

func (r *Room) RoomMessageHandle(session *remote.Session, req request.RoomMessageReq) {
	if req.Type == proto.GetRoomSceneInfoNotify {
		r.GetRoomSceneInfoPush(session)
	}
}

func (r *Room) GetRoomSceneInfoPush(session *remote.Session) {

	userInfoArr := make([]*proto.RoomUser, 0)
	for _, user := range r.users {
		userInfoArr = append(userInfoArr, user)
	}
	data := map[string]any{
		"type":       proto.GetRoomSceneInfoPush,
		"pushRouter": "RoomMessagePush",
		"data": map[string]any{
			"roomID":          r.Id,
			"roomCreatorInfo": r.RoomCreator,
			"gameRule":        r.gameRule,
			"roomUserInfoArr": userInfoArr,
			"gameData":        r.GameFrame.GetGameData(),
		},
	}
	session.Push([]string{session.GetUid()}, data, "ServerMessagePush")
}
