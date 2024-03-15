package article

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
	"strconv"
)

type S3DAO struct {
	oss *s3.Client
	*GORMArticleDao
	bucket *string
}

func NewS3DAO(oss *s3.Client, db *gorm.DB) *S3DAO {
	return &S3DAO{
		oss: oss,
		GORMArticleDao: &GORMArticleDao{
			db: db,
		},
		bucket: aws.String("bucket"),
	}
}

func (dao *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
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
			Content:    "",
			Status:     art.Status,
			AuthorId:   art.AuthorId,
			CreateTime: art.CreateTime,
			UpdateTime: art.UpdateTime,
		})
		if err != nil {
			return err
		}
		// content保存到oss
		_, err = dao.oss.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      dao.bucket,
			Key:         aws.String(strconv.FormatInt(art.Id, 10)),
			Body:        bytes.NewReader([]byte(art.Content)),
			ContentType: aws.String("text/plain;charset=utf-8"),
		})
		return err
	})
	return art.Id, err
}
