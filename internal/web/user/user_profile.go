package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (h *Handler) Profile(ctx *gin.Context) {
	param := ctx.Query("id")
	if param == "" {
		ctx.JSON(http.StatusOK, gin.H{"message": "no param: id"})
		return
	}
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "param is not int64 format"})
		return
	}
	user, err := h.svc.Profile(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": user})
}
