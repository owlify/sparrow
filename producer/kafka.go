package producer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/segmentio/kafka-go"

	"github.com/owlify/sparrow/errors"
)

type KafkaProducerOpts struct {
	Brokers  string
	Topic    string
	MaxRetry int
}

type kafkaProducer struct {
	writer *kafka.Writer
	opts   *KafkaProducerOpts
}

type Producer interface {
	Produce(ctx context.Context, payload interface{}) (err error)
	Close()
}

func NewKafkaProducer(opts *KafkaProducerOpts) Producer {
	return &kafkaProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:     strings.Split(opts.Brokers, ","),
			Topic:       opts.Topic,
			MaxAttempts: opts.MaxRetry,
			BatchSize:   1,
		}),
		opts: opts,
	}
}

func (p *kafkaProducer) Produce(ctx context.Context, payload interface{}) (err error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		err = errors.New("error while unmarshaling payload")
	}

	msg := kafka.Message{
		Value: bytes,
	}

	done := make(chan int)
	go func() {
		err = p.writer.WriteMessages(ctx, msg)
		done <- 0
	}()
	<-done

	return err
}

func (p *kafkaProducer) Close() {
	p.writer.Close()
}
