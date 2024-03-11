package article

import (
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/pkg/ginx/middlewares"
	"net/http"
)

func (h *Handler) Create(ctx *gin.Context) {
	type CreateReq struct {
		Tittle  string `json:"tittle"`
		Content string `json:"content"`
	}
	var req CreateReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	id, err := h.svc.Save(ctx, domain.Article{
		Tittle:   req.Tittle,
		Content:  req.Content,
		AuthorId: ctx.GetInt64(globalkey.JwtUserId),
	})
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, middlewares.Result[int64]{Data: id})
}

func (h *Handler) Edit(ctx *gin.Context) {
	type CreateReq struct {
		Id      int64  `json:"id"`
		Tittle  string `json:"tittle"`
		Content string `json:"content"`
	}
	var req CreateReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	id, err := h.svc.Save(ctx, domain.Article{
		Id:       req.Id,
		Tittle:   req.Tittle,
		Content:  req.Content,
		AuthorId: ctx.GetInt64(globalkey.JwtUserId),
	})
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, middlewares.Result[int64]{Data: id})
}
