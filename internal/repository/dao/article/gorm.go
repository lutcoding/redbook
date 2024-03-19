package article

import (
	"context"
	"errors"
	"fmt"
	"github.com/lutcoding/redbook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

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
			"status":      art.Status,
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
			"status":      art.Status,
			"update_time": now,
		}),
	}).Create(&art).Error
}

func (dao *GORMArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		err = txDAO.Upsert(ctx, PublishArticle{
			Id:         art.Id,
			Tittle:     art.Tittle,
			Content:    art.Content,
			Status:     art.Status,
			AuthorId:   art.AuthorId,
			CreateTime: art.CreateTime,
			UpdateTime: art.UpdateTime,
		})
		return err
	})
	return art.Id, err
}

func (dao *GORMArticleDao) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, authorId).
			Updates(map[string]any{
				"status":      status,
				"update_time": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 1.没有这篇文章
			// 有人攻击网站，试图改写他人文章
			return fmt.Errorf("文章不存在or不是本人文章,无法修改")
		}
		return tx.Model(&PublishArticle{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status":      status,
				"update_time": now,
			}).Error
	})
}

func (dao *GORMArticleDao) GetDraftPageByAuthor(ctx context.Context, uid int64, limit, offset int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Limit(limit).Offset(offset).
		Clauses(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "update_time"}, Desc: true},
		}}).
		Find(&arts).Error

	return arts, err
}

func (dao *GORMArticleDao) GetPubPageByAuthor(ctx context.Context, uid int64, limit, offset int) ([]PublishArticle, error) {
	var arts []PublishArticle
	err := dao.db.WithContext(ctx).
		Where("author_id = ? AND status = ?", uid, domain.ArticleStatusPublished.ToUint8()).
		Limit(limit).Offset(offset).
		Clauses(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: "update_time"}, Desc: true},
			},
		}).Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDao) GetDraftById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}

func (dao *GORMArticleDao) GetPubById(ctx context.Context, id int64) (PublishArticle, error) {
	var art PublishArticle
	err := dao.db.WithContext(ctx).Where("id = ? AND status = ?", id, 2).First(&art).Error
	return art, err
}

func (dao *GORMArticleDao) GetPubPageByTime(ctx context.Context, start time.Time, limit, offset int) ([]PublishArticle, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()
	var arts []PublishArticle
	const ArticleStatusPublished = 2
	err := dao.db.WithContext(ctx).
		Where("update_time < ? AND status = ?", start.UnixMilli(), ArticleStatusPublished).
		Limit(limit).Offset(offset).Find(&arts).Error
	return arts, err
}
