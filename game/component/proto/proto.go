package proto

type GameRule struct {
	AddScores      []int `json:"addScores"`      //加注分
	BaseScore      int   `json:"baseScore"`      //底分
	Bureau         int   `json:"bureau"`         //局数
	CanEnter       bool  `json:"canEnter"`       //中途进人
	CanTrust       bool  `json:"canTrust"`       //允许托管
	CanWatch       bool  `json:"canWatch"`       //允许观战
	Cuopai         bool  `json:"cuopai"`         //高级 是否允许搓牌
	Fangzuobi      bool  `json:"fangzuobi"`      //防作弊
	Yuyin          bool  `json:"yuyin"`          //语音
	GameFrameType  int   `json:"gameFrameType"`  //游戏模式
	GameType       int   `json:"gameType"`       //游戏类型 牛牛 三公等
	MaxPlayerCount int   `json:"maxPlayerCount"` //最大人数
	MinPlayerCount int   `json:"minPlayerCount"` //最小人数
	MaxScore       int   `json:"maxScore"`       //最大加注分
	RoundType      int   `json:"roundType"`      //轮数
	PayDiamond     int   `json:"payDiamond"`     //房费
	PayType        int   `json:"payType"`        //支付方式 1 AA支付 2 赢家支付 3 我支付
	RoomType       int   `json:"roomType"`       // 1 正常房间 2 持续房间 3 百人房间
}

type GameType int
type SendCardType int
type GameFrameType int
type ScaleType int

const (
	PinSanZhang GameType = 1
	NiuNiu               = 2
	PaoDeKuai            = 3
	SanGong              = 4
	HongZhong            = 5
	DouGongNiu           = 8
)
