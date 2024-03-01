package jwt

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"strings"
	"time"
)

func (h *Handler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.NewString()
	err := h.SetAccessJwtToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	session := sessions.Default(ctx)
	session.Set("ssid", ssid)
	session.Options(sessions.Options{
		MaxAge: 60 * 60 * 24 * 7,
	})
	session.Save()
	return h.SetRefreshJwtToken(ctx, uid, ssid)
}

func (h *Handler) SetAccessJwtToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 10)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signedString, err := token.SignedString(h.AccessKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", signedString)
	return nil
}

func (h *Handler) SetRefreshJwtToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signedString, err := token.SignedString(h.RefreshKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", signedString)
	return nil
}

func (h *Handler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

type AccessClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}
