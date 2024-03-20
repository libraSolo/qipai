package handler

import (
	"common"
	"common/biz"
	"common/config"
	"common/jwts"
	"common/logs"
	"connector/models/request"
	"context"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/game"
	"framework/net"
)

type EntryHandler struct {
	userService *service.UserService
}

func NewEntryHandler(r *repo.Manager) *EntryHandler {
	return &EntryHandler{
		userService: service.NewUserService(r),
	}
}

func (h *EntryHandler) Entry(session *net.Session, body []byte) (any, error) {
	logs.Info("entry req params: %v", string(body))
	var req request.EntryReq
	err := json.Unmarshal(body, &req)
	if err != nil {
		return common.F(biz.RequestDataError), nil
	}

	// 校验 token
	uid, err := jwts.ParseToken(req.Token, config.Conf.Jwt.Secret)
	if err != nil {
		logs.Error("parse token error: %v", err)
		return common.F(biz.TokenInfoError), nil
	}
	session.Uid = uid
	// 根据 uid 在 mongo 中查找用户
	user, err := h.userService.FindUserByUid(context.TODO(), uid, req.UserInfo)
	if err != nil {
		return common.F(biz.SqlError), nil
	}
	return common.S(map[string]any{
		"userInfo": user,
		"config":   game.Conf.GetFrontGameConfig(),
	}), nil
}
