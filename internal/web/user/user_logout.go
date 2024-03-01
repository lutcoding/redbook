package user

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Options(sessions.Options{
		MaxAge: -1,
	})
	session.Save()
	ctx.JSON(http.StatusOK, gin.H{"message": "logout success"})
}
