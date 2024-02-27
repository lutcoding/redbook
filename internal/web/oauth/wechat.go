package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/service"
	"github.com/lutcoding/redbook/internal/service/oauth/wechat"
	"net/http"
)

type OAuth2WeChatHandler struct {
	svc     *wechat.Service
	userSvc *service.UserService
}

func NewOAuth2WeChatHandler(svc *wechat.Service, userSvc *service.UserService) *OAuth2WeChatHandler {
	return &OAuth2WeChatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WeChatHandler) AuthURL(ctx *gin.Context) {
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": url})
	return
}

func (h *OAuth2WeChatHandler) CallBack(ctx *gin.Context) {
	code, state := ctx.Query("code"), ctx.Query("state")
	wechatInfo, err := h.svc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	user, err := h.userSvc.FindOrCreateByWeChat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": user.Id})
	return
}

/*type OAuth2Handler struct {
	svc map[string]svc.OAuthService
}

func NewOAuth2Handler() *OAuth2Handler {
	return &OAuth2Handler{}
}

func (h *OAuth2Handler) AuthURL(ctx *gin.Context) {
	platform := ctx.Param("platform")
	h.svc[platform].AuthURL(ctx)
}

oauth2 := unauthorized.Group("/oauth2")
{
	oauth2.Get("/:platform/authurl", s.oauth2Handler.AuthURL)
}*/
