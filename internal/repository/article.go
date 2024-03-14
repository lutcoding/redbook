package repository

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
}

type ArticleCacheRepository struct {
	dao article.ArticleDAO
}

func NewArticleCacheRepository(dao article.ArticleDAO) *ArticleCacheRepository {
	return &ArticleCacheRepository{
		dao: dao,
	}
}

func (repo *ArticleCacheRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.domainToEntity(art))
}

func (repo *ArticleCacheRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.Update(ctx, repo.domainToEntity(art))
}

func (repo *ArticleCacheRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Sync(ctx, repo.domainToEntity(art))
}

func (repo *ArticleCacheRepository) domainToEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Tittle:   art.Tittle,
		Content:  art.Content,
		AuthorId: art.AuthorId,
	}
}
