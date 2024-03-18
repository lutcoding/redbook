package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type ArticleCache interface {
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)

	SetDraft(ctx context.Context, art domain.Article) error
	GetDraft(ctx context.Context, uid int64) (domain.Article, error)

	// SetPub 正常来说，创作者和读者的 Redis 集群要分开，因为读者是一个核心中的核心
	SetPub(ctx context.Context, art domain.Article) error
	DelPub(ctx context.Context, uid int64) error
	GetPub(ctx context.Context, uid int64) (domain.Article, error)
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func NewArticleRedisCache(client redis.Cmdable) *ArticleRedisCache {
	return &ArticleRedisCache{
		client: client,
	}
}

func (cache *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	bytes, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.firstPageKey(uid),
		bytes, time.Minute*10).Err()
}

func (cache *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	bytes, err := cache.client.Get(ctx, cache.firstPageKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bytes, &arts)
	zap.L().Debug("获取缓存草稿首页成功", zap.Int64("user_id", uid))
	return arts, err
}

func (cache *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	return cache.client.Del(ctx, cache.firstPageKey(uid)).Err()
}

func (cache *ArticleRedisCache) SetDraft(ctx context.Context, art domain.Article) error {
	bytes, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.draftKey(art.Id), bytes, time.Minute).Err()
}

func (cache *ArticleRedisCache) GetDraft(ctx context.Context, id int64) (domain.Article, error) {
	bytes, err := cache.client.Get(ctx, cache.draftKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(bytes, &art)
	return art, err
}

func (cache *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.pubKey(art.Id), val, time.Minute*10).Err()
}

func (cache *ArticleRedisCache) DelPub(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, cache.pubKey(id)).Err()
}

func (cache *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := cache.client.Get(ctx, cache.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (cache *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("%v%v", globalkey.PublishedArtCachedPrefix, id)
}

func (cache *ArticleRedisCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("%v%v", globalkey.ArticleFirstCachePage, uid)
}

// 草稿箱的缓存设置
func (cache *ArticleRedisCache) draftKey(id int64) string {
	return fmt.Sprintf("%v%v", globalkey.DraftArtCachePrefix, id)
}
