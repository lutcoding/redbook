package repository

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/cache"
	"github.com/lutcoding/redbook/internal/repository/dao/article"
	"go.uber.org/zap"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error
	ListAuthorDraft(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error)
	ListAuthorPub(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error)
	GetDraft(ctx context.Context, id int64) (domain.Article, error)
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	ListByTime(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error)
	preCache(ctx context.Context, arts []domain.Article)
}

type ArticleCacheRepository struct {
	dao   article.ArticleDAO
	cache cache.ArticleCache
}

func NewArticleCacheRepository(dao article.ArticleDAO, cache cache.ArticleCache) *ArticleCacheRepository {
	return &ArticleCacheRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *ArticleCacheRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.domainToEntity(art))
	if err != nil {
		return 0, err
	}
	go func() {
		err := repo.cache.DelFirstPage(ctx, art.AuthorId)
		if err != nil {
			zap.L().Error("删除缓存redis失败", zap.Error(err))
		}
	}()
	return id, nil
}

func (repo *ArticleCacheRepository) Update(ctx context.Context, art domain.Article) error {
	err := repo.dao.Update(ctx, repo.domainToEntity(art))
	if err != nil {
		return err
	}
	go func() {
		err := repo.cache.DelFirstPage(ctx, art.AuthorId)
		if err != nil {
			zap.L().Error("删除缓存redis失败", zap.Error(err))
		}
	}()
	return nil
}

func (repo *ArticleCacheRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.domainToEntity(art))
	if err != nil {
		return 0, err
	}
	go func() {
		err := repo.cache.DelFirstPage(ctx, art.AuthorId)
		if err != nil {
			zap.L().Error("删除缓存redis失败", zap.Error(err))
		}
	}()
	return id, nil
}

func (repo *ArticleCacheRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error {
	err := repo.dao.SyncStatus(ctx, id, authorId, status.ToUint8())
	if err != nil {
		return err
	}
	go func() {
		err := repo.cache.DelFirstPage(ctx, authorId)
		if err != nil {
			zap.L().Error("删除缓存redis失败", zap.Error(err))
		}
	}()
	return nil
}

func (repo *ArticleCacheRepository) ListAuthorDraft(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error) {
	if 0 <= offset && limit <= 100 {
		arts, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return arts, nil
		}
		zap.L().Debug("查询缓存未命中", zap.Error(err))
	}
	arts, err := repo.dao.GetDraftPageByAuthor(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	fn := func(arts []article.Article) []domain.Article {
		res := make([]domain.Article, len(arts))
		for i, art := range arts {
			res[i] = repo.entityToDraftDomain(art)
		}
		return res
	}
	res := fn(arts)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := repo.cache.SetFirstPage(ctx, uid, res)
		if err != nil {
			zap.L().Error("设置缓存redis失败", zap.Error(err))
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.preCache(ctx, res)
	}()
	return res, nil
}

func (repo *ArticleCacheRepository) ListAuthorPub(ctx context.Context, uid int64, limit, offset int) ([]domain.Article, error) {
	arts, err := repo.dao.GetPubPageByAuthor(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	fn := func(arts []article.PublishArticle) []domain.Article {
		res := make([]domain.Article, len(arts))
		for i, art := range arts {
			res[i] = repo.entityToPubDomain(art)
		}
		return res
	}
	return fn(arts), nil
}

func (repo *ArticleCacheRepository) GetDraft(ctx context.Context, id int64) (domain.Article, error) {
	art, err := repo.dao.GetDraftById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.entityToDraftDomain(art), nil
}

func (repo *ArticleCacheRepository) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.entityToPubDomain(art), nil
}

func (repo *ArticleCacheRepository) ListByTime(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error) {
	arts, err := repo.dao.GetPubPageByTime(ctx, time.Now(), limit, offset)
	if err != nil {
		return nil, err
	}
	fn := func(arts []article.PublishArticle) []domain.Article {
		res := make([]domain.Article, len(arts))
		for i, art := range arts {
			res[i] = repo.entityToPubDomain(art)
		}
		return res
	}
	return fn(arts), nil
}

func (repo *ArticleCacheRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := repo.cache.SetDraft(ctx, arts[0])
		if err != nil {
			zap.L().Error("删除缓存redis失败", zap.Error(err))
		}
	}
}

func (repo *ArticleCacheRepository) domainToEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Tittle:   art.Tittle,
		Content:  art.Content,
		AuthorId: art.AuthorId,
		Status:   art.ArticleStatus.ToUint8(),
	}
}

func (repo *ArticleCacheRepository) entityToDraftDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:            art.Id,
		Tittle:        art.Tittle,
		Content:       art.Content,
		AuthorId:      art.AuthorId,
		ArticleStatus: domain.ArticleStatus(art.Status),
	}
}

func (repo *ArticleCacheRepository) entityToPubDomain(art article.PublishArticle) domain.Article {
	return domain.Article{
		Id:            art.Id,
		Tittle:        art.Tittle,
		Content:       art.Content,
		AuthorId:      art.AuthorId,
		ArticleStatus: domain.ArticleStatus(art.Status),
	}
}
