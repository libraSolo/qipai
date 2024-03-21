package request

import "game/component/proto"

type CreateRoomReq struct {
	UnionID    int64          `json:"unionID"`
	GameRuleID string         `json:"gameRuleID"`
	GameRule   proto.GameRule `json:"gameRule"`
}
