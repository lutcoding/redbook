package dao

import (
	"context"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const (
	statusWait = iota
	statusSuccess
	statusFail
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSmsDAO interface {
	Insert(ctx context.Context, sms AsyncSms) error
	GetWaitingSms(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type AsyncSmsGormDAO struct {
	db *gorm.DB
}

func NewAsyncSmsGormDAO(db *gorm.DB) *AsyncSmsGormDAO {
	return &AsyncSmsGormDAO{
		db: db,
	}
}

func (dao *AsyncSmsGormDAO) Insert(ctx context.Context, sms AsyncSms) error {
	now := time.Now().UnixMilli()
	sms.CreateTime, sms.UpdateTime, sms.Status = now, now, statusWait
	return dao.db.WithContext(ctx).Create(&sms).Error
}

func (dao *AsyncSmsGormDAO) GetWaitingSms(ctx context.Context) (AsyncSms, error) {
	var sms AsyncSms
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("update_time < ? AND status = ?", endTime, statusWait).
			First(&sms).Error
		if err != nil {
			return err
		}

		err = tx.Model(&AsyncSms{}).
			Where("id = ?", sms.Id).
			Updates(map[string]any{
				"retry_cnt":   gorm.Expr("retry_cnt + ?", 1),
				"update_time": now,
			}).Error
		return err
	})
	return sms, err
}

func (dao *AsyncSmsGormDAO) MarkSuccess(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      statusSuccess,
			"update_time": time.Now().UnixMilli(),
		}).Error
}

func (dao *AsyncSmsGormDAO) MarkFailed(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ? AND `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"status":      statusFail,
			"update_time": time.Now().UnixMilli(),
		}).Error
}

type AsyncSms struct {
	Id             int64 `gorm:"primaryKey, autoIncrement"`
	RetryCnt       int64
	RetryMax       int64
	AsyncSmsConfig datatypes.JSON
	Status         int
	CreateTime     int64
	UpdateTime     int64
}
