package common

import (
	"common/biz"
	"framework/errorCode"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Result struct {
	Code int `json:"code"`
	Msg  any `json:"msg"`
}

func Fail(ctx *gin.Context, err *errorCode.Error) {
	ctx.JSON(http.StatusOK, Result{
		Code: err.Code,
		Msg:  err.Err.Error(),
	})
}

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Result{
		Code: biz.OK,
		Msg:  data,
	})
}
