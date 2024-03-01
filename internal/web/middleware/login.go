package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lutcoding/redbook/common/globalkey"
	jwtHdl "github.com/lutcoding/redbook/internal/web/jwt"
	"net/http"
)

type LoginMiddlewareBuilder struct {
	jwtHdl *jwtHdl.Handler
}

func NewLoginMiddlewareBuilder(jwtHdl *jwtHdl.Handler) *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{
		jwtHdl: jwtHdl,
	}
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := l.jwtHdl.ExtractToken(ctx)
		claims := &jwtHdl.AccessClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return l.jwtHdl.AccessKey, nil
		})
		if err != nil || !token.Valid || claims.Uid == 0 || claims.UserAgent != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		session := sessions.Default(ctx)
		v := session.Get("ssid")
		if v == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set(globalkey.JwtUserId, claims.Uid)
	}
}
