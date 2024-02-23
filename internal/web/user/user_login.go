package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/service"
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
	signedString := h.getJwtToken(ctx, user)
	ctx.Header("x-jwt-token", signedString)
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

func (h *Handler) getJwtToken(ctx *gin.Context, user domain.User) string {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.Id,
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signedString, _ := token.SignedString([]byte("NqdHZfporsLtXRTPhc01IZJXDnFsaTHsmsMWixjPEgQJyiZxsXKcsmkg1XvAWXIp"))
	return signedString
}

type Claims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
