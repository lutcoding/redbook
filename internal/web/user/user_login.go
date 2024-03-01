package user

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/service"
	jwtHdl "github.com/lutcoding/redbook/internal/web/jwt"
	"net/http"
)

func (h *Handler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}

	var user domain.User
	user, err = h.svc.Login(ctx, domain.User{Email: req.Email, Password: req.Password})
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmailOrPassword) {
			ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		} else {
			ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
			return
		}
	}
	if h.jwtHdl.SetLoginToken(ctx, user.Id) != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "login success"})
	return
}

func (h *Handler) SendLoginSmsCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	// TODO: 校验手机号是否合法
	err = h.smsSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "send verification code success"})
	return
}

func (h *Handler) LoginSmsCode(ctx *gin.Context) {
	type LoginReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	// TODO: 校验格式
	err = h.smsSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "verify code success"})
	return
}

func (h *Handler) Refresh(ctx *gin.Context) {
	tokenStr := h.jwtHdl.ExtractToken(ctx)
	claims := &jwtHdl.RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return h.jwtHdl.RefreshKey, nil
	})
	if err != nil || !token.Valid || claims.Uid == 0 || claims.UserAgent != ctx.Request.UserAgent() {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	session := sessions.Default(ctx)
	v := session.Get("ssid")
	if v == nil {
		fmt.Println("2")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.jwtHdl.SetLoginToken(ctx, claims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
