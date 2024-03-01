package user

import (
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/common/globalkey"
	"net/http"
)

func (h *Handler) Profile(ctx *gin.Context) {
	value, exists := ctx.Get(globalkey.JwtUserId)
	if !exists {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	user, err := h.svc.Profile(ctx, value.(int64))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": user})
}
