package handler

import (
	"common"
	"common/biz"
	"context"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/remote"
	"game/logic"
	"game/models/request"
)

type UnionHandler struct {
	um          *logic.UnionManager
	userService *service.UserService
}

func NewUnionHandler(r *repo.Manager, manager *logic.UnionManager) *UnionHandler {
	return &UnionHandler{
		um:          manager,
		userService: service.NewUserService(r),
	}
}

func (h *UnionHandler) JoinRoom(session *remote.Session, msg []byte) any {
	// 判断 uid 是否合法
	uid := session.GetUid()
	if len(uid) <= 0 {
		return common.F(biz.InvalidUsers)
	}
	var req request.JoinRoomReq
	if err := json.Unmarshal(msg, &req); err != nil {
		return common.F(biz.RequestDataError)
	}
	// 判断用户是否存在
	user, err := h.userService.FindUserByUid(context.TODO(), uid)
	if err != nil {
		return common.F(biz.SqlError)
	}
	if user == nil {
		return common.F(biz.NotFindUser)
	}
	// 获取房间
	bizErr := h.um.JoinRoom(session, req.RoomID, user)
	if bizErr != nil {
		return common.F(bizErr)
	}
	return common.S(nil)
}

func (h *UnionHandler) CreateRoom(session *remote.Session, msg []byte) any {
	// 判断 uid 是否合法
	uid := session.GetUid()
	if len(uid) <= 0 {
		return common.F(biz.InvalidUsers)
	}
	var req request.CreateRoomReq
	if err := json.Unmarshal(msg, &req); err != nil {
		return common.F(biz.RequestDataError)
	}
	// 判断用户是否存在
	user, err := h.userService.FindUserByUid(context.TODO(), uid)
	if err != nil {
		return common.F(biz.SqlError)
	}
	if user == nil {
		return common.F(biz.NotFindUser)
	}
	// TODO 判断session中是否存在roomID，判断是否在房间中
	// 创建房间
	union := h.um.GetUnion(req.UnionID)
	if union == nil {
		return common.F(biz.UnionNotExist)
	}
	bizErr := union.CreateRoom(h.userService, session, req, user)
	if bizErr != nil {
		return common.F(bizErr)
	}

	return common.S(nil)
}
