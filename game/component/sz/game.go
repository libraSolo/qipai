package sz

import (
	"common/logs"
	"common/utils"
	"encoding/json"
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
	// 庄家的为首次操作号
	g.gameData.CurChairID = g.gameData.BankerChairID

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
		// chairID 是做操作的座次号
		g.ServerMessagePush(session, []string{roomUser.UserInfo.Uid}, GameTurnPushData(g.gameData.CurChairID, g.gameData.CurScore))
	}
}

func (g *GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	// 1. 解析参数
	var req MessageReq
	json.Unmarshal(msg, &req)
	// 2. 根据不同类型触发不同的操作
	if req.Type == GameLookNotify {
		g.onGameLook(user, session, req.Data.Cuopai)
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

func (g *GameFrame) IsPlayingChairID(chairID int) bool {
	for _, v := range g.r.GetUsers() {
		if v.ChairId == chairID && v.UserStatus == proto.Playing {
			return true
		}
	}
	return false
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

func (g *GameFrame) onGameLook(user *proto.RoomUser, session *remote.Session, cuopai bool) {
	// 判断 如果是当前用户 推送其牌, 给其他用户推送此用户看牌状态
	if g.gameData.GameStatus != PourScore || g.gameData.CurChairID != user.ChairId {
		logs.Warn("ID:%s room, sanZhang game look err:gameStatus=%d, curChairID=%d, chairID=%d",
			g.r.GetID(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairId)
		return
	}
	if !g.IsPlayingChairID(user.ChairId) {
		logs.Warn("ID:%s room, sanZhang game look err:gameStatus=%d, curChairID=%d, chairID=%d",
			g.r.GetID(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairId)
		return
	}
	// 已经看牌
	g.gameData.UserStatusArray[user.ChairId] = Look
	g.gameData.LookCards[user.ChairId] = 1
	for _, v := range g.r.GetUsers() {
		// 当前用户
		if g.gameData.CurChairID == v.ChairId {
			g.ServerMessagePush(session, []string{v.UserInfo.Uid}, GameLookPushData(g.gameData.CurChairID, g.gameData.HandCards[v.ChairId], cuopai))
			continue
		}
		g.ServerMessagePush(session, []string{v.UserInfo.Uid}, GameLookPushData(g.gameData.CurChairID, nil, cuopai))

	}

}
