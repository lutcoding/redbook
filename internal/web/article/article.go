package article

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/pkg/ginx/middlewares"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) Create(ctx *gin.Context) {
	type CreateReq struct {
		Tittle  string `json:"tittle"`
		Content string `json:"content"`
	}
	var req CreateReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
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
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
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

func (h *Handler) GetDraft(ctx *gin.Context) {
	type GetDraftReq struct {
		Id int64 `json:"id"`
	}
	var req GetDraftReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
		return
	}
	art, err := h.svc.GetDraft(ctx, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, middlewares.Result[domain.Article]{Data: art})
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
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
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
		Status   uint8  `json:"status"`
	}
	var req ListReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
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

func (h *Handler) ListPub(ctx *gin.Context) {
	type ListReq struct {
		Uid    int64 `json:"uid"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
	}
	type ArticleVO struct {
		Id       int64  `json:"id"`
		Tittle   string `json:"tittle"`
		Abstract string `json:"abstract"`
		Status   uint8  `json:"status"`
	}
	var req ListReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
		return
	}
	arts, err := h.svc.ListAuthorPub(ctx, req.Uid, req.Limit, req.Offset)
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
				Abstract: art.Abstract(),
				Status:   art.ArticleStatus.ToUint8(),
			}
		}
		return res
	}
	ctx.JSON(http.StatusOK, middlewares.Result[[]ArticleVO]{
		Data: fn(arts),
	})
}

func (h *Handler) GetPub(ctx *gin.Context) {
	param := ctx.Param("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: err.Error()})
		return
	}
	type ArticleVO struct {
		Art   domain.Article
		Inter domain.Interactive
	}
	var res ArticleVO
	var eg errgroup.Group

	// TODO 这里有BUG 这篇文章对应表interactives中如果没有记录, 使用errgroup.Group就会返回错误
	eg.Go(func() error {
		art, err := h.svc.GetPub(ctx, id)
		if err != nil {
			return err
		}
		res.Art = art
		return nil
	})
	value, exists := ctx.Get(globalkey.JwtUserId)
	if !exists {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "server internal error"})
		return
	}
	uid := value.(int64)
	eg.Go(func() error {
		info, err := h.interSvc.GetInteractiveInfo(ctx, uid, h.biz, id)
		if err != nil {
			return err
		}
		res.Inter = info
		return nil
	})

	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "server internal error"})
		return
	}

	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := h.interSvc.IncrReadCnt(newCtx, h.biz, id)
		if err != nil {
			zap.L().Error("设置缓存失败", zap.Error(err))
		}
	}()
	fmt.Println(res)
	ctx.JSON(http.StatusOK, middlewares.Result[ArticleVO]{Data: res})
}

func (h *Handler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "解析json错误，请传入正确参数"})
		return
	}
	value, exists := ctx.Get(globalkey.JwtUserId)
	if !exists {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "server internal error"})
		return
	}
	uid := value.(int64)
	if req.Like {
		err = h.interSvc.Like(ctx, uid, h.biz, req.Id)
	} else {
		err = h.interSvc.CancelLike(ctx, uid, h.biz, req.Id)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, middlewares.Result[int64]{Msg: "ok"})
}
