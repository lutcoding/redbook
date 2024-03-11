package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/web/jwt"
	"net/http"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T, uc jwt.AccessClaims) (Result[T], error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		err := ctx.Bind(&req)
		if err != nil {
			return
		}
		result, err := fn(ctx, req, jwt.AccessClaims{})
		if err != nil {
			// 统一日志处理
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

type Result[T any] struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
