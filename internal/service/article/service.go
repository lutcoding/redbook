package article

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/events/article"
	"github.com/lutcoding/redbook/internal/repository"
)

type Service struct {
	repo     repository.ArticleRepository
	producer article.Producer
}

func NewService(repo repository.ArticleRepository, producer article.Producer) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
	}
}

func (s *Service) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetPub(ctx, id)
}

func (s *Service) GetDraft(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetDraft(ctx, id)
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

func (s *Service) ListAuthorPub(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error) {
	return s.repo.ListAuthorPub(ctx, uid, limit, offset)
}
