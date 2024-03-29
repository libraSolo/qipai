package request

import "game/component/proto"

type RoomMessageReq struct {
	Type proto.RoomMessageType `json:"type"`
	Data map[string]any        `json:"data"`
}
