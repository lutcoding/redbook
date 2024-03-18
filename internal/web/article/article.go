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

func (h *Handler) Publish(ctx *gin.Context) {
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
	id, err := h.svc.Sync(ctx, domain.Article{
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

func (h *Handler) ToPrivate(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	id := ctx.GetInt64(globalkey.JwtUserId)
	// 健全逻辑, 系统错误
	if id == 0 {

	}
	err = h.svc.ToPrivate(ctx, req.Id, id)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "ok"})
}

func (h *Handler) ListDraft(ctx *gin.Context) {
	type ListReq struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}
	type ArticleVO struct {
		Id       int64  `json:"id"`
		Tittle   string `json:"tittle"`
		Abstract string `json:"abstract"`
		Content  string `json:"content"`
		Status   uint8  `json:"status"`
	}
	var req ListReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: err.Error()})
		return
	}
	articles, err := h.svc.ListAuthorDraft(ctx, ctx.GetInt64(globalkey.JwtUserId), req.Limit, req.Offset)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: err.Error()})
		return
	}
	fn := func(arts []domain.Article) []ArticleVO {
		res := make([]ArticleVO, len(arts))
		for i, art := range arts {
			res[i] = ArticleVO{
				Id:       art.Id,
				Tittle:   art.Tittle,
				Content:  art.Content,
				Abstract: art.Abstract(),
				Status:   art.ArticleStatus.ToUint8(),
			}
		}
		return res
	}
	ctx.JSON(http.StatusOK, middlewares.Result[[]ArticleVO]{
		Data: fn(articles),
	})
}
