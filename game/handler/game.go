package handler

import (
	"common"
	"common/biz"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/remote"
	"game/logic"
	"game/models/request"
)

type GameHandler struct {
	um          *logic.UnionManager
	userService *service.UserService
}

func NewGameHandler(r *repo.Manager, manager *logic.UnionManager) *GameHandler {
	return &GameHandler{
		um:          manager,
		userService: service.NewUserService(r),
	}
}

func (h *GameHandler) RoomMessageNotify(session *remote.Session, msg []byte) any {
	if len(session.GetUid()) <= 0 {
		return common.F(biz.InvalidUsers)
	}
	var req request.RoomMessageReq
	if err := json.Unmarshal(msg, &req); err != nil {
		return common.F(biz.RequestDataError)
	}
	roomId, ok := session.Get("roomId")
	if !ok {
		return common.F(biz.RoomNotExist)
	}
	rm := h.um.GetRoomById(roomId.(string))
	if rm == nil {
		return common.F(biz.RoomNotExist)
	}
	rm.RoomMessageHandle(session, req)
	return nil
}

func (h *GameHandler) GameMessageNotify(session *remote.Session, msg []byte) any {
	if len(session.GetUid()) <= 0 {
		return common.F(biz.InvalidUsers)
	}

	roomId, ok := session.Get("roomId")
	if !ok {
		return common.F(biz.RoomNotExist)
	}
	rm := h.um.GetRoomById(roomId.(string))
	if rm == nil {
		return common.F(biz.RoomNotExist)
	}
	rm.GameMessageHandle(session, msg)
	return nil
}
