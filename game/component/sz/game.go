package sz

import (
	"common/logs"
	"common/utils"
	"encoding/json"
	"framework/remote"
	"game/component/base"
	"game/component/proto"
	"github.com/jinzhu/copier"
	"time"
)

type GameFrame struct {
	r          base.RoomFrame
	gameRule   proto.GameRule
	gameData   *GameData
	logic      *Logic
	gameResult *GameResult
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
	g.Winner = make([]int, 0)
	return g
}

func (g *GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	// 1. 解析参数
	var req MessageReq
	json.Unmarshal(msg, &req)
	// 2. 根据不同类型触发不同的操作
	switch req.Type {
	case GameLookNotify:
		g.onGameLook(user, session, req.Data.Cuopai)
	case GamePourScoreNotify:
		g.onGamePourScore(user, session, req.Data.Score, req.Data.Type)
	case GameCompareNotify:
		g.onGameCompare(user, session, req.Data.ChairID)
	default:
		logs.Info("没有匹配的操作类型 %d", req.Type)
	}
}

func (g *GameFrame) ServerMessagePush(session *remote.Session, users []string, data any) {
	session.Push(users, data, "ServerMessagePush")
}

func (g *GameFrame) GetGameData(session *remote.Session) any {
	// 获取当前用户
	user := g.r.GetUsers()[session.GetUid()]
	// 判断当前用户是否已经看牌
	// 已经看牌则返回, 否则为隐藏
	var gameData GameData
	_ = copier.CopyWithOption(&gameData, g.gameData, copier.Option{DeepCopy: true})
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.gameData.HandCards[i] != nil {
			// 隐藏
			gameData.HandCards[i] = make([]int, 3)
		} else {
			// 没有玩的人
			gameData.HandCards[i] = nil
		}
	}
	if g.gameData.LookCards[user.ChairId] == 1 {
		// 已经看牌
		gameData.HandCards[user.ChairId] = g.gameData.HandCards[user.ChairId]
	}

	return gameData
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
		g.ServerMessagePush(session, []string{roomUser.UserInfo.Uid}, GamePourScorePushData(roomUser.ChairId, g.gameData.CurScore, g.gameData.CurScore, 1, 0))
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

func (g *GameFrame) onGamePourScore(user *proto.RoomUser, session *remote.Session, score int, t int) {
	// 1. 处理下分 保存用户下的分数 同时推送当前用户下分的信息
	if g.gameData.GameStatus != PourScore || g.gameData.CurChairID != user.ChairId {
		logs.Warn("ID:%s room, sanZhang game gamePourScore err:gameStatus=%d, curChairID=%d, chairID=%d",
			g.r.GetID(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairId)
		return
	}
	// 不在游戏状态
	if !g.IsPlayingChairID(user.ChairId) {
		logs.Warn("ID:%s room, sanZhang game gamePourScore err:gameStatus=%d, curChairID=%d, chairID=%d",
			g.r.GetID(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairId)
		return
	}
	if score < 0 {
		logs.Warn("ID:%s room, sanZhang game gamePourScore err:gameStatus=%d, curChairID=%d, chairID=%d",
			g.r.GetID(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairId)
		return
	}
	if g.gameData.PourScores[user.ChairId] == nil {
		g.gameData.PourScores[user.ChairId] = make([]int, 0)
	}
	g.gameData.PourScores[user.ChairId] = append(g.gameData.PourScores[user.ChairId], score)
	// 所有人分数
	scores := 0
	for i := 0; i < g.gameData.ChairCount; i++ {
		// 没分的人
		if g.gameData.PourScores[i] == nil {
			continue
		}
		for _, sc := range g.gameData.PourScores[i] {
			scores += sc
		}
	}
	// 当前座次的总分
	curScores := 0
	for _, sc := range g.gameData.PourScores[user.ChairId] {
		curScores += sc
	}

	g.ServerMessagePush(session, g.getAllUsers(), GamePourScorePushData(user.ChairId, score, curScores, scores, t))
	// 2. 结束下分 座次移动到下一位 推送轮次、游戏状态、操作的座次
	g.endPourScore(session)
}

func (g *GameFrame) endPourScore(session *remote.Session) {
	// 1. 推送轮数 轮数大于规则的限制 结束游戏 进行结算
	round := g.getCurRound()
	g.ServerMessagePush(session, g.getAllUsers(), GameRoundPushData(round))
	// 判断当前的玩家没有输掉 只剩下一个
	gamerCount := 0
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.IsPlayingChairID(i) && !utils.Contains(g.gameData.Loser, i) {
			gamerCount++
		}
	}
	if gamerCount == 1 {
		// 只剩下一个玩家
		g.startResult(session)
		return
	}
	// 非一个玩家 继续
	// 2. 座次要向前移动一位
	for i := 0; i < g.gameData.ChairCount; i++ {
		g.gameData.CurChairID++
		g.gameData.CurChairID = g.gameData.CurChairID % g.gameData.ChairCount
		if g.IsPlayingChairID(i) {
			break
		}
	}
	// 推送状态
	// 推送下分状态
	g.gameData.GameStatus = PourScore
	g.ServerMessagePush(session, g.getAllUsers(), GameStatusPushData(g.gameData.GameStatus, 30))
	// 推操作
	g.ServerMessagePush(session, g.getAllUsers(), GameTurnPushData(g.gameData.CurChairID, g.gameData.CurScore))

}

func (g *GameFrame) getCurRound() int {
	// 作为并非从0开始
	cur := g.gameData.CurChairID
	for i := 0; i < g.gameData.ChairCount; i++ {
		cur++
		cur = cur % g.gameData.ChairCount
		if g.IsPlayingChairID(cur) {
			return len(g.gameData.PourScores[cur])
		}
	}
	return 0
}

func (g *GameFrame) onGameCompare(user *proto.RoomUser, session *remote.Session, otherChairID int) {
	// 1.TODO: 先下分 跟注结束后 进行比牌
	// 2. 比牌
	curChairID := user.ChairId
	result := g.logic.CompareCards(g.gameData.HandCards[curChairID], g.gameData.HandCards[otherChairID])
	// 3. 处理比牌结果 推送轮数 状态 显示结果等信息
	//if result == 0 {
	//	// 平局 主动比牌者输
	//	result = -1
	//}
	// 默认 当前用户赢
	winChairID := curChairID
	loseChairID := otherChairID
	if result <= 0 {
		// 平局 主动比牌者输
		winChairID = otherChairID
		loseChairID = curChairID
	}
	g.ServerMessagePush(session, g.getAllUsers(), GameComparePushData(curChairID, otherChairID, winChairID, loseChairID))
	g.gameData.UserStatusArray[winChairID] = Win
	g.gameData.Winner = append(g.gameData.Winner, winChairID)

	g.gameData.UserStatusArray[loseChairID] = Lose
	g.gameData.Loser = append(g.gameData.Loser, loseChairID)
	//TODO 赢了后 继续和其他人进行比牌
	g.endPourScore(session)
}

func (g *GameFrame) startResult(session *remote.Session) {
	// 推送 游戏结束
	g.gameData.GameStatus = Result
	g.ServerMessagePush(session, g.getAllUsers(), GameStatusPushData(g.gameData.GameStatus, 0))

	// 推送比赛结果
	if g.gameResult == nil {
		g.gameResult = new(GameResult)
	}
	g.gameResult.Winners = g.gameData.Winner
	g.gameResult.HandCards = g.gameData.HandCards
	g.gameResult.CurScores = g.gameData.CurScores
	g.gameResult.Losers = g.gameData.Loser

	winScores := make([]int, g.gameData.ChairCount)
	for i := range winScores {
		if g.gameData.PourScores[i] != nil {
			scores := 0
			for _, v := range g.gameData.PourScores[i] {
				scores += v
			}
			winScores[i] = -scores

			for win := range g.gameData.Winner {
				winScores[win] += scores / len(g.gameData.Winner)
			}
		}
	}
	g.gameResult.WinScores = winScores
	g.ServerMessagePush(session, g.getAllUsers(), GameResultPushData(g.gameResult))
	// 结算完成 重置游戏 开始下一把
	g.resetGame(session)
	g.gameEnd(session)
}

func (g *GameFrame) resetGame(session *remote.Session) {
	newData := &GameData{
		GameType:   GameType(g.gameRule.GameFrameType),
		BaseScore:  g.gameRule.BaseScore,
		ChairCount: g.gameRule.MaxPlayerCount,
	}
	newData.PourScores = make([][]int, newData.ChairCount)
	newData.HandCards = make([][]int, newData.ChairCount)
	newData.LookCards = make([]int, newData.ChairCount)
	newData.CurScores = make([]int, 0)
	newData.UserStatusArray = make([]UserStatus, newData.ChairCount)
	newData.UserTrustArray = make([]bool, 10)
	newData.Loser = make([]int, 0)
	newData.Winner = make([]int, 0)
	newData.GameStatus = GameStatus(0)
	g.gameData = newData
	g.SendGameStatus(session, g.gameData.GameStatus, 0)
	// 房间重置
	g.r.EndGame(session)
}

func (g *GameFrame) SendGameStatus(session *remote.Session, status GameStatus, tick int) {
	g.ServerMessagePush(session, g.getAllUsers(), GameStatusPushData(status, tick))
}

func (g *GameFrame) gameEnd(session *remote.Session) {
	// 赢家当庄
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.gameResult.WinScores[i] > 0 {
			g.gameData.BankerChairID = i
			g.gameData.CurChairID = i
		}
	}
	time.AfterFunc(5*time.Second, func() {
		for _, user := range g.r.GetUsers() {
			g.r.UserReady(session, user.UserInfo.Uid)
		}
	})
}
