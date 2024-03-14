package article

import "context"

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoArticleDao struct {
	client *mongo.Client
}

func NewMangoArticleDao() *MongoArticleDao {
	return &MongoArticleDao{}
}

func (m *MongoArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) Update(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) Upsert(ctx context.Context, art PublishArticle) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}
