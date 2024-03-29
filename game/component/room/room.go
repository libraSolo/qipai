package room

import (
	"common/logs"
	"core/models/entity"
	"framework/remote"
	"game/component/base"
	"game/component/proto"
	"game/component/sz"
	"game/models/request"
	"sync"
	"time"
)

type Room struct {
	sync.RWMutex
	Id            string
	unionID       int64
	gameRule      proto.GameRule
	users         map[string]*proto.RoomUser
	RoomCreator   *proto.RoomCreator
	GameFrame     GameFrame
	kickSchedules map[string]*time.Timer
	union         base.UnionBase
	roomDismissed bool
}

func NewRoom(id string, unionID int64, rule proto.GameRule, u base.UnionBase) *Room {
	room := &Room{
		Id:            id,
		unionID:       unionID,
		gameRule:      rule,
		users:         make(map[string]*proto.RoomUser),
		kickSchedules: make(map[string]*time.Timer),
		union:         u,
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
	// 超时踢出用户
	go r.addTickKickScheduleEvent(session, user.Uid)
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

func (r *Room) ServerMessagePush(session *remote.Session, user []string, data any) {
	session.Push(user, data, "ServerMessagePush")
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

func (r *Room) addTickKickScheduleEvent(session *remote.Session, uid string) {
	r.kickSchedules[uid] = time.AfterFunc(30*time.Second, func() {
		logs.Info("kick... user not ready")
		// 取消任务 (timer.afterFunc 单次执行, 不取消问题不大)
		timer, ok := r.kickSchedules[uid]
		if ok {
			timer.Stop()
		}
		delete(r.kickSchedules, uid)
		// 判断用户的状态
		user, ok := r.users[uid]
		if ok {
			if user.UserStatus < proto.Ready {
				r.kickUser(user, session)
				// 判断是否解散房间
				if len(r.users) == 0 {
					r.dismissRoom()
				}
			}
		}
	})
}

func (r *Room) kickUser(user *proto.RoomUser, session *remote.Session) {
	// 房间id 为空, 则是踢出
	r.ServerMessagePush(session, []string{user.UserInfo.Uid}, proto.UpdateUserInfoPush(""))
	// 通知其他用户离开房间
	users := make([]string, 0)
	for _, v := range r.users {
		users = append(users, v.UserInfo.Uid)
	}
	r.ServerMessagePush(session, users, proto.UserLeaveRoomPushData(user))
	delete(r.users, user.UserInfo.Uid)
}

func (r *Room) dismissRoom() {
	r.Lock()
	defer r.Unlock()
	if r.roomDismissed {
		return
	}
	r.roomDismissed = true
	r.cancelAllScheduler()
	r.union.DismissRoom(r.Id)
}

func (r *Room) cancelAllScheduler() {
	// 取消房间的所有定时任务
	for uid, timer := range r.kickSchedules {
		timer.Stop()
		delete(r.kickSchedules, uid)
	}

}
