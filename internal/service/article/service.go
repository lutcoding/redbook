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

func (s *Service) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.ArticleStatus = domain.ArticleStatusUnPublished
	if art.Id == 0 {
		return s.Create(ctx, art)
	}
	return s.Update(ctx, art)
}

func (s *Service) Create(ctx context.Context, art domain.Article) (int64, error) {
	return s.repo.Create(ctx, art)
}

func (s *Service) Update(ctx context.Context, art domain.Article) (int64, error) {
	return art.Id, s.repo.Update(ctx, art)
}

func (s *Service) Sync(ctx context.Context, art domain.Article) (int64, error) {
	art.ArticleStatus = domain.ArticleStatusPublished
	return s.repo.Sync(ctx, art)
}

func (s *Service) ToPrivate(ctx context.Context, id int64, authorId int64) error {
	return s.repo.SyncStatus(ctx, id, authorId, domain.ArticleStatusPrivate)
}

func (s *Service) ListAuthorDraft(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error) {
	return s.repo.ListAuthorDraft(ctx, uid, limit, offset)
}
