package dao

import (
	"github.com/lutcoding/redbook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &AsyncSms{},
		&article.Article{}, &article.PublishArticle{}, &Interactive{}, &LikeInfo{})
}
