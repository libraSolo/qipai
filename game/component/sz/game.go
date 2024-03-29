package sz

import (
	"game/component/base"
	"game/component/proto"
)

type GameFrame struct {
	r        base.RoomFrame
	gameRule proto.GameRule
	GameData *GameData
}

func NewGameFrame(r base.RoomFrame, rule proto.GameRule) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{r: r, gameRule: rule, GameData: gameData}
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

func (g *GameFrame) GetGameData() any {
	return g.GameData
}
