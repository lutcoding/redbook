package article

import (
	"github.com/lutcoding/redbook/internal/service"
	"github.com/lutcoding/redbook/internal/service/article"
)

type Handler struct {
	svc      *article.Service
	interSvc *service.InteractiveService
	biz      string
}

func NewHandler(svc *article.Service, interSvc *service.InteractiveService) *Handler {
	return &Handler{
		svc:      svc,
		interSvc: interSvc,
		biz:      "article",
	}
}
