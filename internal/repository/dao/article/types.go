package article

import (
	"context"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	Update(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
	GetDraftPageByAuthor(ctx context.Context, uid int64, limit, offset int) ([]Article, error)
	GetPubPageByAuthor(ctx context.Context, uid int64, limit, offset int) ([]PublishArticle, error)
	GetDraftById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishArticle, error)
	GetPubPageByTime(ctx context.Context, start time.Time, limit, offset int) ([]PublishArticle, error)
}
