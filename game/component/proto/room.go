package proto

import "core/models/entity"

type RoomCreator struct {
	Uid         string     `json:"uid"`
	CreatorType CreateType `json:"creatorType"`
}

type CreateType int

const (
	UserCreatorType CreateType = iota + 1
	UnionCreatorType
)

type UserInfo struct {
	Uid          string `json:"uid"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Gold         int64  `json:"gold"`
	FrontendId   string `json:"frontendId"`
	Address      string `json:"address"`
	Location     string `json:"location"`
	LastLoginIP  string `json:"lastLoginIP"`
	Sex          int    `json:"sex"`
	Score        int    `json:"score"`
	SpreaderID   string `json:"spreaderID"` //推广ID
	ProhibitGame bool   `json:"prohibitGame"`
	RoomID       string `json:"roomID"`
}

type UserStatus int

const (
	None UserStatus = 1 << iota
	Ready
	Playing
	Offline
	Dismiss
)

type RoomUser struct {
	UserInfo   UserInfo   `json:"userInfo"`
	ChairId    int        `json:"chairID"`
	UserStatus UserStatus `json:"userStatus"`
}

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
