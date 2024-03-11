package article

import "github.com/lutcoding/redbook/internal/service/article"

type Handler struct {
	svc *article.Service
}

func NewHandler(svc *article.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}
