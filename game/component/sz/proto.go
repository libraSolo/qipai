package sz

type MessageData struct {
	Cuopai  bool `json:"cuopai"`
	Score   int  `json:"score"`
	Type    int  `json:"type"` // 1:跟注 2:加注
	ChairID int  `json:"chairID"`
}

type MessageReq struct {
	Type int         `json:"type"`
	Data MessageData `json:"data"`
}
type GameStatus int

type GameData struct {
	BankerChairID   int                      `json:"bankerChairID"` // 庄家
	ChairCount      int                      `json:"chairCount"`
	CurBureau       int                      `json:"curBureau"` // 局数
	CurScore        int                      `json:"curScore"`  // 当前分数
	CurScores       []int                    `json:"curScores"`
	GameStarter     bool                     `json:"gameStarter"`
	GameStatus      GameStatus               `json:"gameStatus"`
	HandCards       [][]int                  `json:"handCards"` // 手牌
	LookCards       []int                    `json:"lookCards"`
	Loser           []int                    `json:"loser"`
	Winner          []int                    `json:"winner"`
	MaxBureau       int                      `json:"maxBureau"`
	PourScores      [][]int                  `json:"pourScores"`
	GameType        GameType                 `json:"gameType"`
	BaseScore       int                      `json:"baseScore"`
	Result          any                      `json:"result"`
	Round           int                      `json:"round"`
	Tick            int                      `json:"tick"` //倒计时
	UserTrustArray  []bool                   `json:"userTrustArray"`
	UserStatusArray []UserStatus             `json:"userStatusArray"`
	UserWinRecord   map[string]UserWinRecord `json:"userWinRecord"`
	ReviewRecord    []BureauReview           `json:"reviewRecord"`
	TrustTmArray    []int                    `json:"trustTmArray"`
	CurChairID      int                      `json:"curChairID"`
}

const (
	GameStatusPush      = 401 //游戏状态推送
	GameSendCardsPush   = 402 //发牌推送
	GameLookNotify      = 303 //看牌请求
	GameLookPush        = 403
	GamePourScoreNotify = 304 //下分请求
	GamePourScorePush   = 404
	GameCompareNotify   = 305 //比牌请求
	GameComparePush     = 405
	GameTurnPush        = 406 //操作推送
	GameResultPush      = 407 //结果推送
	GameEndPush         = 409 //结束推送
	GameChatNotify      = 310 //游戏聊天
	GameChatPush        = 410
	GameBureauPush      = 411 //局数推送
	GameAbandonNotify   = 312 //弃牌请求
	GameAbandonPush     = 412
	GameRoundPush       = 413 //轮数推送
	GameBankerPush      = 414 //庄家推送
	GameTrustNotify     = 315 //托管
	GameTrustPush       = 415 //托管推送
	GameReviewNotify    = 316 //牌面回顾
	GameReviewPush      = 416
)

// None 初始状态
const None int = 0
const (
	SendCards GameStatus = 1 //发牌中
	PourScore            = 2 //下分中
	Result               = 3 //显示结果
)

type GameStatusTm int

const (
	TmSendCards GameStatusTm = 1
	TmPourScore              = 30 //下分中
	TmResult                 = 5  //显示结果
)

type GameType int

const (
	Men1 GameType = 1 //闷1轮
	Men2          = 2 //闷2轮
	Men3          = 3 //闷3轮
)

type RoundType int

const (
	Round10 RoundType = 1 //10轮
	Round20           = 2 //15轮
	Round30           = 3 //20轮
)

type CardsType int

const (
	DanZhang CardsType = 1 //单牌
	DuiZi              = 2 //对子
	ShunZi             = 3 //顺子
	JinHua             = 4 //金花
	ShunJin            = 5 //顺金
	BaoZi              = 6 //豹子
)

type UserStatus int

const (
	Abandon        UserStatus = 1 << iota // 放弃
	TimeoutAbandon                        //超时放弃
	Look                                  //看牌
	Lose                                  //比牌失败
	Win                                   //胜利
	He                                    //和
)

type UserWinRecord struct {
	Uid      string `json:"uid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Score    int    `json:"score"`
}

type BureauReview struct {
	Uid       string `json:"uid"`
	Cards     []int  `json:"cards"`
	PourScore int    `json:"pourScore"`
	WinScore  int    `json:"winScore"`
	NickName  string `json:"nickName"`
	Avatar    string `json:"avatar"`
	IsBanker  bool   `json:"isBanker"`
	IsAbandon bool   `json:"isAbandon"`
}

type GameResult struct {
	Winners   []int   `json:"winners"`
	WinScores []int   `json:"winScores"`
	HandCards [][]int `json:"handCards"`
	CurScores []int   `json:"curScores"`
	Losers    []int   `bson:"losers"`
}
