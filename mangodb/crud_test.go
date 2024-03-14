package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMango(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command.String())
		},
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {

		},
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {

		},
	}
	ops := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetMonitor(monitor).
		SetAuth(options.Credential{Username: "root", Password: "123456"})
	client, err := mongo.Connect(ctx, ops)
	assert.NoError(t, err)
	database := client.Database("redbook")
	collection := database.Collection("articles")
	result, err := collection.InsertMany(ctx, []interface{}{
		Article{
			Id:      123,
			Tittle:  "this is tittle",
			Content: "this is content",
		},
	})
	assert.NoError(t, err)
	fmt.Printf("id: %v\n", result.InsertedIDs)

	var art Article
	err = collection.FindOne(ctx, Article{Id: 123}).Decode(&art)
	assert.NoError(t, err)

	filter := bson.D{bson.E{Key: "id", Value: 123}}
	update := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "tittle", Value: "new tittle"}}}
	updateResult, err := collection.UpdateMany(ctx, filter, update)
	assert.NoError(t, err)
	collection.UpdateMany(ctx, filter, bson.D{bson.E{Key: "$set", Value: Article{Tittle: "new tittle"}}})
	fmt.Println(updateResult.ModifiedCount)
}

type Article struct {
	Id      int64  `bson:"id,omitempty"`
	Tittle  string `bson:"tittle,omitempty"`
	Content string `bson:"content,omitempty"`

	Status   uint8 `bson:"status,omitempty"`
	AuthorId int64 `bson:"author_id,omitempty"`

	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
}
