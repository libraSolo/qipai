package router

import (
	"common/config"
	"common/rpc"
	handler "gate/handle"
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册路由
func RegisterRouter() *gin.Engine {
	if config.Conf.Log.Level == "DEBUG" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// gate 作为 grpc 的客户端，调用 user grpc 服务
	rpc.Init()

	r := gin.Default()
	userHandler := handler.NewUserHandler()
	r.POST("/register", userHandler.Register)
	return r
}
