package mj

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

func GameBankerPushData(bankerChairID int) any {
	return map[string]any{
		"type": GameBankerPush,
		"data": map[string]any{
			"bankerChairID": bankerChairID,
		},
		"pushRouter": "GameMessagePush",
	}
}

func GameDicelPushData(d1, d2 int) any {
	return map[string]any{
		"type": GameDicesPush,
		"data": map[string]any{
			"dice1": d1,
			"dice2": d2,
		},
		"pushRouter": "GameMessagePush",
	}
}
