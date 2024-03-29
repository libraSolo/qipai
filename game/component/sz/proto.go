package sz

type GameStatus int

type GameData struct {
	BankerChairID   int                      `json:"bankerChairID"`
	ChairCount      int                      `json:"chairCount"`
	CurBureau       int                      `json:"curBureau"`
	CurScore        int                      `json:"curScore"`
	CurScores       []int                    `json:"curScores"`
	GameStarter     bool                     `json:"gameStarter"`
	GameStatus      GameStatus               `json:"gameStatus"`
	HandCards       [][]int                  `json:"handCards"`
	LookCards       []int                    `json:"lookCards"`
	Loser           []int                    `json:"loser"`
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
	JinHua             = 3 //金花
	ShunJin            = 3 //顺金
	BaoZi              = 3 //豹子
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
