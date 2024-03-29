package proto

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
	None  UserStatus = 0
	Ready UserStatus = 1 << (iota - 1)
	Playing
	Offline
	Dismiss
)

type RoomUser struct {
	UserInfo   UserInfo   `json:"userInfo"`
	ChairId    int        `json:"chairID"`
	UserStatus UserStatus `json:"userStatus"`
}
