package article

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	Update(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
}

type GORMArticleDao struct {
	db *gorm.DB
}

func NewGORMArticleDao(db *gorm.DB) *GORMArticleDao {
	return &GORMArticleDao{
		db: db,
	}
}

func (dao *GORMArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime, art.UpdateTime = now, now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDao) Update(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.Model(&art).WithContext(ctx).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"tittle":      art.Tittle,
			"content":     art.Content,
			"update_time": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败")
	}
	return nil
}

func (dao *GORMArticleDao) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.CreateTime, art.UpdateTime = now, now
	return dao.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"tittle":      art.Tittle,
			"content":     art.Content,
			"update_time": now,
		}),
	}).Create(&art).Error
}

func (dao *GORMArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var (
			id  = art.Id
			err error
		)
		txDAO := NewGORMArticleDao(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.Update(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		err = txDAO.Upsert(ctx, PublishArticle{art})
		return err
	})
	return art.Id, err
}

type Article struct {
	Id      int64  `gorm:"primaryKey, autoIncrement"`
	Tittle  string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`

	AuthorId int64

	CreateTime int64
	UpdateTime int64
}

type PublishArticle struct {
	Article
}
