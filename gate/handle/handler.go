package handler

import (
	"common"
	"common/biz"
	"common/config"
	"common/jwts"
	"common/logs"
	"common/rpc"
	"context"
	"framework/errorCode"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) Register(ctx *gin.Context) {
	// 接收参数
	var req pb.RegisterParams
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		logs.Error("get req error: %v", err)
		common.Fail(ctx, biz.RequestDataError)
		return
	}
	res, err := rpc.UserClient.Register(context.TODO(), &req)
	if err != nil {
		logs.Error("grpc failed err: %v", err)
		common.Fail(ctx, errorCode.ToError(err))
		return
	}
	uid := res.Uid
	if len(uid) == 0 {
		common.Fail(ctx, biz.RequestDataError)
		return
	}
	logs.Info("uid:%s", uid)
	// gen token by uid
	claims := jwts.CustomClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	token, err := jwts.GenToken(&claims, config.Conf.Jwt.Secret)
	if err != nil {
		logs.Error("jwt gen token failed err: %v", err)
		common.Fail(ctx, biz.Fail)
		return
	}
	result := map[string]any{
		"token": token,
		"serverInfo": map[string]any{
			"host": config.Conf.Services["connector"].ClientHost,
			"port": config.Conf.Services["connector"].ClientPort,
		},
	}
	common.Success(ctx, result)
}
