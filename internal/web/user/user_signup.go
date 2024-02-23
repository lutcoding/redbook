package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/domain"
)

func (h *Handler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, gin.H{"message": "Password not equal ConfirmPassword"})
		return
	}
	ok, err := h.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "server internal error"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, gin.H{"message": "email format error"})
		return
	}
	err = h.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "sign up success"})
}
