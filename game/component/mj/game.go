package mj

import (
	"common/utils"
	"framework/remote"
	"game/component/base"
	"game/component/proto"
	"github.com/jinzhu/copier"
)

type GameFrame struct {
	r        base.RoomFrame
	gameRule proto.GameRule
	gameData *GameData
	logic    *Logic
}

func (g *GameFrame) GetGameData(session *remote.Session) any {
	// 获取场景 获取游戏数据
	chairID := g.r.GetUsers()[session.GetUid()].ChairId
	var gameData GameData
	_ = copier.CopyWithOption(&gameData, g.gameData, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	handCards := make([][]CardID, g.gameData.ChairCount)
	for i := range gameData.HandCards {
		if i == chairID {
			handCards[i] = gameData.HandCards[i]
		} else {
			// 非自己 每个牌变为 36(翻过来的牌)
			handCards[i] = make([]CardID, len(g.gameData.HandCards[i]), 36)
			//handCards[i] = make([]int, len(g.gameData.HandCards[i]))
			//for j := range g.gameData.HandCards[i] {
			//	handCards[i][j] = 36
			//}
		}
	}
	gameData.HandCards = handCards

	if g.gameData.GameStatus == GameStatusNone {
		gameData.RestCardsCount = 9*3*4 + 4
		if g.gameRule.GameFrameType == HongZhong8 {
			gameData.RestCardsCount = 9*3*4 + 8
		}
	}
	return g.gameData
}

func (g *GameFrame) StartGame(session *remote.Session, user *proto.RoomUser) {
	// 开始游戏
	// 1.游戏状态 初始状态 推送
	g.gameData.GameStarted = true
	g.gameData.GameStatus = Dices
	g.sendData(session, GameStatusPushData(g.gameData.GameStatus, GameStatusTmDices))
	// 2.庄家推送
	if g.gameData.CurBureau == 0 {
		g.gameData.BankerChairID = 0
	} else {
		// 胜利者是庄
	}
	g.sendData(session, GameBankerPushData(g.gameData.BankerChairID))
	// 3.摇骰子推送
	dice1 := utils.Rand(6) + 1
	dice2 := utils.Rand(6) + 1
	g.sendData(session, GameDicelPushData(dice1, dice2))
	// 4.发牌推送
	g.sendHandCards(session)
	// 10.当前局数推送
}

func (g *GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	//TODO implement me
	panic("implement me")
}

func (g *GameFrame) sendHandCards(session *remote.Session) {
	// 洗牌
	g.logic.washCards()
	// 发牌
	for i := 0; i < g.gameData.ChairCount; i++ {
		// 每个人13张牌
		g.gameData.HandCards[i] = g.logic.getCards(13)
	}
	// 5.剩余牌数推送
	// 6.局数推送
	// 7.开始游戏状态推送
	// 8.拿牌推送
	// 9.剩余牌数推送
}

func NewGameFrame(r base.RoomFrame, rule proto.GameRule) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{r: r,
		gameRule: rule,
		gameData: gameData,
		logic:    NewLogic(GameType(rule.GameFrameType), rule.Qidui)}
}

func initGameData(rule proto.GameRule) *GameData {
	g := &GameData{
		ChairCount:     rule.MaxPlayerCount,
		UserTrustArray: make([]int, 0),
		HandCards:      make([][]CardID, rule.MaxPlayerCount),
		GameStatus:     GameStatusNone,
		OperateRecord:  make([]OperateRecord, 0),
		OperateArrays:  make([][]OperateType, 0),
		CurChairID:     -1,
	}
	g.RestCardsCount = 9*3*4 + 4
	if rule.GameFrameType == HongZhong8 {
		g.RestCardsCount = 9*3*4 + 8
	}

	return g
}
func (g *GameFrame) sendDataUsers(session *remote.Session, users []string, data any) {
	g.ServerMessagePush(session, users, data)
}

func (g *GameFrame) sendData(session *remote.Session, data any) {
	g.ServerMessagePush(session, g.getAllUsers(), data)
}

func (g *GameFrame) ServerMessagePush(session *remote.Session, users []string, data any) {
	session.Push(users, data, "ServerMessagePush")
}

func (g *GameFrame) getAllUsers() []string {
	users := make([]string, 0)
	for _, v := range g.r.GetUsers() {
		users = append(users, v.UserInfo.Uid)
	}
	return users
}
