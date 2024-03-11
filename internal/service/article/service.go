package article

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository"
)

type Service struct {
	repo repository.ArticleRepository
}

func NewService(repo repository.ArticleRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Save(ctx context.Context, article domain.Article) (int64, error) {
	if article.Id == 0 {
		return s.Create(ctx, article)
	}
	return s.Update(ctx, article)
}

func (s *Service) Create(ctx context.Context, article domain.Article) (int64, error) {
	return s.repo.Create(ctx, article)
}

func (s *Service) Update(ctx context.Context, article domain.Article) (int64, error) {
	return article.Id, s.repo.Update(ctx, article)
}
