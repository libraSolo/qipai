package sz

import (
	"common/utils"
	"framework/remote"
	"game/component/base"
	"game/component/proto"
)

type GameFrame struct {
	r        base.RoomFrame
	gameRule proto.GameRule
	gameData *GameData
	logic    *Logic
}

func NewGameFrame(r base.RoomFrame, rule proto.GameRule) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{r: r, gameRule: rule, gameData: gameData, logic: NewLogic()}
}

func initGameData(rule proto.GameRule) *GameData {
	g := &GameData{
		GameType:   GameType(rule.GameFrameType),
		BaseScore:  rule.BaseScore,
		ChairCount: rule.MaxPlayerCount,
	}
	g.PourScores = make([][]int, g.ChairCount)
	g.HandCards = make([][]int, g.ChairCount)
	g.LookCards = make([]int, g.ChairCount)
	g.CurScores = make([]int, 0)
	g.UserStatusArray = make([]UserStatus, g.ChairCount)
	g.UserTrustArray = make([]bool, 10)
	g.Loser = make([]int, 0)
	return g
}

func (g *GameFrame) ServerMessagePush(session *remote.Session, users []string, data any) {
	session.Push(users, data, "ServerMessagePush")
}

func (g *GameFrame) GetGameData() any {
	return g.gameData
}

func (g *GameFrame) StartGame(session *remote.Session, user *proto.RoomUser) {
	// 1.用户信息变更推送（金币变化） {"gold": 9958, "pushRouter": 'UpdateUserInfoPush'}
	users := g.getAllUsers()
	g.ServerMessagePush(session, users, UpdateUserInfoPushGold(user.UserInfo.Gold))
	// 2.庄家推送 {"type":414,"data":{"bankerChairID":0},"pushRouter":"GameMessagePush"}
	if g.gameData.CurBureau == 0 {
		// 随机庄家
		g.gameData.BankerChairID = utils.Rand(len(users))
	}
	g.ServerMessagePush(session, users, GameBankerPushData(g.gameData.BankerChairID))
	// 3.局数推送{"type":411,"data":{"curBureau":6},"pushRouter":"GameMessagePush"}
	g.gameData.CurBureau++
	g.ServerMessagePush(session, users, GameBureauPushData(g.gameData.CurBureau))
	// 4.游戏状态推送 分两步推送 第一步 推送 发牌 牌发完之后 第二步 推送下分 需要用户操作了 推送操作
	// 发牌 {"type":401,"data":{"gameStatus":1,"tick":0},"pushRouter":"GameMessagePush"}
	g.gameData.GameStatus = SendCards
	g.ServerMessagePush(session, users, GameStatusPushData(g.gameData.GameStatus, 0))
	// 发牌推送
	g.sendCards(session)
	// 推送下分状态
	g.gameData.GameStatus = PourScore
	g.ServerMessagePush(session, users, GameStatusPushData(g.gameData.GameStatus, 30))
	// 推送下分
	g.gameData.CurScore = g.gameRule.AddScores[0] * g.gameRule.BaseScore
	for _, roomUser := range g.r.GetUsers() {
		g.ServerMessagePush(session, []string{roomUser.UserInfo.Uid}, GamePourScorePushData(roomUser.ChairId, g.gameData.CurScore, g.gameData.CurScore, 1))
	}
	// 轮数推送
	g.gameData.Round = 1
	g.ServerMessagePush(session, users, GameRoundPushData(g.gameData.Round))
	// 操作推送
	for _, roomUser := range g.r.GetUsers() {
		g.ServerMessagePush(session, []string{roomUser.UserInfo.Uid}, GameTurnPushData(roomUser.ChairId, g.gameData.CurScore))
	}
}

func (g *GameFrame) getAllUsers() []string {
	users := make([]string, 0)
	for _, v := range g.r.GetUsers() {
		users = append(users, v.UserInfo.Uid)
	}
	return users
}

func (g *GameFrame) sendCards(session *remote.Session) {
	// 洗牌 发牌
	g.logic.washCards()
	playingUsers := g.getPlayingUsers()
	for i := 0; i < len(playingUsers); i++ {
		g.gameData.HandCards[i] = g.logic.getCards()
	}
	// 发牌后, 不看牌则是暗牌
	hands := make([][]int, g.gameData.ChairCount)
	for i, cards := range g.gameData.HandCards {
		if cards != nil {
			hands[i] = []int{0, 0, 0}
		}
	}
	g.ServerMessagePush(session, g.getAllUsers(), GameSendCardsPushData(hands))

}

func (g *GameFrame) getPlayingUsers() []string {
	users := make([]string, 0)
	for _, v := range g.r.GetUsers() {
		if v.UserStatus == proto.Playing {
			users = append(users, v.UserInfo.Uid)
		}
	}
	return users
}
