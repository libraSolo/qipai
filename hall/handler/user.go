package handler

import (
	"common"
	"common/biz"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/remote"
	"hall/models/request"
	"hall/models/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(r *repo.Manager) *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(r),
	}
}

func (h *UserHandler) UpdateUserAddress(session *remote.Session, msg []byte) any {
	//logs.Info("updateuseraddress msg:%v", string(msg))
	var req request.UpdateUserAddressReq
	if err := json.Unmarshal(msg, &req); err != nil {
		return common.F(biz.RequestDataError)
	}
	err := h.userService.UpdateUserAddressByUid(session.GetUid(), req)
	if err != nil {
		return common.F(biz.SqlError)
	}
	res := &response.UpdateUserAddressRes{
		Result: common.Result{
			Code: biz.OK,
		},
		UpdateUserData: req,
	}

	return res
}
