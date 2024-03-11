package repository

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}

type ArticleCacheRepository struct {
	dao dao.ArticleDAO
}

func NewArticleCacheRepository(dao dao.ArticleDAO) *ArticleCacheRepository {
	return &ArticleCacheRepository{
		dao: dao,
	}
}

func (r *ArticleCacheRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return r.dao.Insert(ctx, r.domainToEntity(article))
}

func (r *ArticleCacheRepository) Update(ctx context.Context, article domain.Article) error {
	return r.dao.Update(ctx, r.domainToEntity(article))
}

func (r *ArticleCacheRepository) domainToEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Tittle:   article.Tittle,
		Content:  article.Content,
		AuthorId: article.AuthorId,
	}
}
