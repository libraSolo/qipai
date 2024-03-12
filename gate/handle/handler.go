package handler

import (
	"common/logs"
	"common/rpc"
	"context"
	"github.com/gin-gonic/gin"
	"user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) Register(ctx *gin.Context) {
	res, err := rpc.UserClient.Register(context.TODO(), &pb.RegisterParams{
		Account:       "",
		Password:      "",
		LoginPlatform: 0,
		SmsCode:       "",
	})
	if err != nil {
		// deal error
	}
	uid := res.Uid
	logs.Info("uid:%s", uid)
	// gen token by uid
	ctx.JSON(200, map[string]interface{}{})
}
