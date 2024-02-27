package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("user has already exists")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	Update(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (u User, err error)
	FindById(ctx context.Context, id int64) (u User, err error)
	FindByPhone(ctx context.Context, phone string) (u User, err error)
	FindByWeChat(ctx context.Context, openID string) (u User, err error)
}

type UserGormDAO struct {
	db *gorm.DB
}

func NewUserGormDAO(db *gorm.DB) UserDAO {
	return &UserGormDAO{db}
}

func (dao *UserGormDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.CreateTime, u.UpdateTime = now, now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		// mysql唯一索引错误码
		const uniqueConflictsErrNo = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *UserGormDAO) Update(ctx context.Context, u User) error {
	u.UpdateTime = time.Now().UnixMilli()
	return dao.db.Model(&u).Updates(u).Error
}

func (dao *UserGormDAO) FindByEmail(ctx context.Context, email string) (u User, err error) {
	err = dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return
}

func (dao *UserGormDAO) FindById(ctx context.Context, id int64) (u User, err error) {
	err = dao.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return
}

func (dao *UserGormDAO) FindByPhone(ctx context.Context, phone string) (u User, err error) {
	err = dao.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return
}

func (dao *UserGormDAO) FindByWeChat(ctx context.Context, openID string) (u User, err error) {
	err = dao.db.WithContext(ctx).First(&u, "wechat_open_id = ?", openID).Error
	return
}

// User model
type User struct {
	Id       int64          `gorm:"primaryKey, autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string

	WechatOpenID  sql.NullString `gorm:"unique"`
	WechatUnionID sql.NullString

	CreateTime int64
	UpdateTime int64
}
