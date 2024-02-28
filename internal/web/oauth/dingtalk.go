package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/service"
	"github.com/lutcoding/redbook/internal/service/oauth/dingtalk"
	"net/http"
)

type OAuth2DingTalkHandler struct {
	svc     *dingtalk.Service
	userSvc *service.UserService
}

func NewOAuth2DingTalkHandler(svc *dingtalk.Service, userSvc *service.UserService) *OAuth2DingTalkHandler {
	return &OAuth2DingTalkHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2DingTalkHandler) AuthURL(ctx *gin.Context) {
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": url})
	return
}

func (h *OAuth2DingTalkHandler) CallBack(ctx *gin.Context) {
	if ctx.Query("error") != "" {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	authCode, state := ctx.Query("authCode"), ctx.Query("state")
	err := h.svc.VerifyCode(ctx, authCode, state)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
