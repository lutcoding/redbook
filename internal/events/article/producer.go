package article

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, event ReadEvent) error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, event ReadEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.ByteEncoder(bytes),
	})
	return err
}

func NewKafkaProducer(producer sarama.SyncProducer, topic string) *KafkaProducer {
	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}
}

type ReadEvent struct {
	Uid int64
	Aid int64
}
