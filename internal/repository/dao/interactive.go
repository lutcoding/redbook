package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrDuplicateLike  = errors.New("重复点赞")
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrCollectCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) error
	DelLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) error
	GetLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) (LikeInfo, error)
	GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (Interactive, error)
}

var (
	ErrDislikeNoRow = errors.New("无法取消点赞")
)

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func (dao *GORMInteractiveDAO) GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz_id = ? AND biz = ?", bizId, biz).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) (LikeInfo, error) {
	var info LikeInfo
	err := dao.db.WithContext(ctx).
		Where("uid = ? AND biz_id = ? AND biz = ? AND status = ?", uid, bizId, biz, 1).
		First(&info).Error
	return info, err
}

func (dao *GORMInteractiveDAO) DelLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&LikeInfo{}).
			Where("uid = ? AND biz_id = ? AND biz = ? AND status = ?", uid, bizId, biz, 1).
			Updates(map[string]any{
				"status":      2,
				"update_time": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrDislikeNoRow
		}
		return tx.Model(&Interactive{}).
			Where("biz_id = ? AND biz = ?", bizId, biz).
			Updates(map[string]any{
				"like_cnt":    gorm.Expr("like_cnt - 1"),
				"update_time": now,
			}).Error
	})
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, uid int64, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status":      1,
				"update_time": now,
			}),
		}).Create(&LikeInfo{
			Uid:        uid,
			Biz:        biz,
			BizId:      bizId,
			Status:     1,
			CreateTime: now,
			UpdateTime: now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrDuplicateLike
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt":    gorm.Expr("`like_cnt` + 1"),
				"update_time": now,
			}),
		}).Create(&Interactive{
			BizId:      bizId,
			Biz:        biz,
			LikeCnt:    1,
			CreateTime: now,
			UpdateTime: now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) IncrCollectCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"update_time": now,
			}),
		}).Create(&Interactive{
		BizId:      bizId,
		Biz:        biz,
		CollectCnt: 1,
		CreateTime: now,
		UpdateTime: now,
	}).Error
}

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"read_cnt":    gorm.Expr("`read_cnt` + 1"),
				"update_time": now,
			}),
		}).Create(&Interactive{
		BizId:      bizId,
		Biz:        biz,
		ReadCnt:    1,
		CreateTime: now,
		UpdateTime: now,
	}).Error
}

func NewGORMInteractiveDAO(db *gorm.DB) *GORMInteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizid, biz>
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	CreateTime int64
	UpdateTime int64
}

type LikeInfo struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 1:未删除  2:删除
	Status     uint8
	CreateTime int64
	UpdateTime int64
}
