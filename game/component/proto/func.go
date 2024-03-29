package proto

import "core/models/entity"

func ToRoomUser(user *entity.User) *UserInfo {
	return &UserInfo{
		Uid:         user.Uid,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Address:     user.Address,
		Location:    user.Location,
		LastLoginIP: user.LastLoginIp,
		Gold:        user.Gold,
		Sex:         user.Sex,
		FrontendId:  user.FrontendId,
	}
}

func UpdateUserInfoPush(roomId string) any {
	pushMsg := map[string]any{
		"roomID":     roomId,
		"pushRouter": "UpdateUserInfoPush",
	}
	return pushMsg
}

func UserLeaveRoomPushData(roomUserInfo *RoomUser) any {
	pushMsg := map[string]any{
		"type": UserLeaveRoomPush,
		"data": map[string]any{
			"roomUserInfo": roomUserInfo,
		},
		"pushRouter": "RoomMessagePush",
	}
	return pushMsg
}
