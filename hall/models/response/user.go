package response

import (
	"common"
	"hall/models/request"
)

type UpdateUserAddressRes struct {
	common.Result
	UpdateUserData request.UpdateUserAddressReq `json:"updateUserData"`
}
