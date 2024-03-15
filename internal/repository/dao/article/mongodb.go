package article

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoArticleDao struct {
	client *mongo.Client
	col    *mongo.Collection
	// 代表的是线上库
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func NewMangoArticleDao(db *mongo.Database, node *snowflake.Node) *MongoArticleDao {
	return &MongoArticleDao{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func (dao *MongoArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateTime, art.UpdateTime = now, now
	art.Id = int64(dao.node.Generate())
	_, err := dao.col.InsertOne(ctx, art)
	return art.Id, err
}

func (dao *MongoArticleDao) Update(ctx context.Context, art Article) error {
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":       art.Tittle,
		"content":     art.Content,
		"update_time": time.Now().UnixMilli(),
		"status":      art.Status,
	}}}
	res, err := dao.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 这边就是校验了 author_id 是不是正确的 ID
	if res.ModifiedCount == 0 {
		return fmt.Errorf("更新数据失败")
	}
	return nil
}

func (dao *MongoArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	panic("")
}

func (dao *MongoArticleDao) Upsert(ctx context.Context, art PublishArticle) error {
	//TODO implement me
	panic("implement me")
}

func (dao *MongoArticleDao) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "create_time", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}
