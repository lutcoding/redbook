package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	Update(ctx context.Context, article Article) error
}

type GORMArticleDao struct {
	db *gorm.DB
}

func NewGORMArticleDao(db *gorm.DB) *GORMArticleDao {
	return &GORMArticleDao{
		db: db,
	}
}

func (dao *GORMArticleDao) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.CreateTime, article.UpdateTime = now, now
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (dao *GORMArticleDao) Update(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.Model(&article).WithContext(ctx).
		Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"tittle":      article.Tittle,
			"content":     article.Content,
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

type Article struct {
	Id      int64  `gorm:"primaryKey, autoIncrement"`
	Tittle  string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`

	AuthorId int64

	CreateTime int64
	UpdateTime int64
}
