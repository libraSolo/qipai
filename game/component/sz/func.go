package sz

// UpdateUserInfoPushGold {"gold": 9958, "pushRouter": 'UpdateUserInfoPush'}
func UpdateUserInfoPushGold(gold int64) any {
	return map[string]any{
		"gold":       gold,
		"pushRouter": "UpdateUserInfoPush",
	}
}

// GameBankerPushData 2.庄家推送 {"type":414,"data":{"bankerChairID":0},"pushRouter":"GameMessagePush"}
func GameBankerPushData(bankerChairID int) any {
	return map[string]any{
		"type": GamePourScorePush,
		"data": map[string]any{
			"bankerChairID": bankerChairID,
		},
		"pushRouter": "GameMessagePush",
	}
}

// GameBureauPushData 3.局数推送{"type":411,"data":{"curBureau":6},"pushRouter":"GameMessagePush"}
func GameBureauPushData(curBureau int) any {
	return map[string]any{
		"type": GameBureauPush,
		"data": map[string]any{
			"curBureau": curBureau,
		},
		"pushRouter": "GameMessagePush",
	}
}

// GameStatusPushData {"type":401,"data":{"gameStatus":1,"tick":0},"pushRouter":"GameMessagePush"}
func GameStatusPushData(status GameStatus, tick int) any {
	return map[string]any{
		"type": GameStatusPush,
		"data": map[string]any{
			"gameStatus": status,
			"tick":       tick,
		},
		"pushRouter": "GameMessagePush",
	}
}

// GameSendCardsPushData 发牌推送
func GameSendCardsPushData(handCards [][]int) any {
	return map[string]any{
		"type": GameSendCardsPush,
		"data": map[string]any{
			"handCards": handCards,
		},
		"pushRouter": "GameMessagePush",
	}
}

// GamePourScorePushData 座次, 玩家拥有分数, 当前座次所下分数, 所有用户下的分数,
func GamePourScorePushData(chairID, score, chairScore, scores, t int) any {
	return map[string]any{
		"type": GamePourScorePush,
		"data": map[string]any{
			"chairID":    chairID,
			"score":      score,
			"chairScore": chairScore,
			"scores":     scores,
			"type":       t,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameRoundPushData(round int) any {
	return map[string]any{
		"type": GameRoundPush,
		"data": map[string]any{
			"round": round,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameTurnPushData(chairID, score int) any {
	return map[string]any{
		"type": GameTurnPush,
		"data": map[string]any{
			"curChairID": chairID,
			"curScore":   score,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameLookPushData(chairID int, cards []int, cuoPai bool) any {
	return map[string]any{
		"type": GameLookPush,
		"data": map[string]any{
			"cards":   cards,
			"chairID": chairID,
			"cuopai":  cuoPai,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameComparePushData(curChairID, otherChairID, winChairID, loseChairID int) any {
	return map[string]any{
		"type": GameComparePush,
		"data": map[string]any{
			"fromChairID": curChairID,
			"toChairID":   otherChairID,
			"winChairID":  winChairID,
			"loseChairID": loseChairID,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameResultPushData(result *GameResult) any {
	return map[string]any{
		"type": GameResultPush,
		"data": map[string]any{
			"result": result,
		},
		"pushRouter": "GameMessagePush",
	}
}
