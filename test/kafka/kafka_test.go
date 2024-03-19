package test

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, cfg)
	assert.NoError(t, err)
	producer, err := sarama.NewSyncProducerFromClient(client)
	assert.NoError(t, err)
	defer producer.Close()
	message, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test",
		Value: sarama.StringEncoder("this is a message1"),
		Headers: []sarama.RecordHeader{
			{Key: []byte("head"), Value: []byte("head_value")},
		},
	})
	assert.NoError(t, err)
	t.Log(message, offset)
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	client, err := sarama.NewClient([]string{"localhost:9094"}, cfg)
	assert.NoError(t, err)
	producer, err := sarama.NewAsyncProducerFromClient(client)
	assert.NoError(t, err)
	input := producer.Input()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case errMsg := <-producer.Errors():
			t.Log(errMsg.Err.Error())
		case <-producer.Successes():
			t.Log("success")
		}
		cancel()
	}()
	input <- &sarama.ProducerMessage{
		Topic: "test",
		Value: sarama.StringEncoder("message1"),
		Headers: []sarama.RecordHeader{
			{Key: []byte("head"), Value: []byte("head_value")},
		},
	}
	<-ctx.Done()
}
