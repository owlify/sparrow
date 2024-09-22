package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime/debug"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"github.com/owlify/sparrow/errors"
	"github.com/owlify/sparrow/logger"
)

type KafkaConsumerOpts struct {
	Brokers    string
	GroupID    string
	Topic      string
	MinBytes   int
	MaxBytes   int
	MaxRetry   int
	SASLConfig *KafkaSASLOpts
}

type KafkaSASLOpts struct {
	Username string
	Password string
}

type kafkaConsumer struct {
	reader  *kafka.Reader
	opts    *KafkaConsumerOpts
	handler EventHandler
}

type Consumer interface {
	Start(ctx context.Context)
	RegisterHandler(handler EventHandler)
	Close()
}

const (
	maxBackoff = time.Second * 12
	minOffset  = time.Millisecond * 400
	maxJitter  = time.Millisecond * 800
)

func NewKafkaConsumer(opts *KafkaConsumerOpts) Consumer {
	var dialer *kafka.Dialer
	if opts.SASLConfig != nil {
		dialer = &kafka.Dialer{
			SASLMechanism: plain.Mechanism{
				Username: opts.SASLConfig.Username,
				Password: opts.SASLConfig.Password,
			},
		}
	} else {
		dialer = &kafka.Dialer{} // Default dialer without SASL
	}

	return &kafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        strings.Split(opts.Brokers, ","),
			GroupID:        opts.GroupID,
			Topic:          opts.Topic,
			MinBytes:       opts.MinBytes,
			MaxBytes:       opts.MaxBytes,
			CommitInterval: 0, // no auto commit
			StartOffset:    kafka.LastOffset,
			Dialer:         dialer,
		}),
		opts: opts,
	}
}

func recoverConsumerPanic(ctx context.Context) {
	if r := recover(); r != nil {
		errorMessage := fmt.Sprintf("%v", r)
		err := errors.New("error while consuming event")
		logger.E(ctx, err, "[KafkaConsumer] error while consuming event",
			logger.Field("error", errorMessage),
			logger.Field("stacktrace", string(debug.Stack())))
	}
}

func (c *kafkaConsumer) Start(ctx context.Context) {
	for {
		c.consume(ctx)
	}
}

func (c *kafkaConsumer) RegisterHandler(handler EventHandler) {
	c.handler = handler
}

func (c *kafkaConsumer) consume(ctx context.Context) {
	defer recoverConsumerPanic(ctx)

	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		logger.E(ctx, err, "[KafkaConsumer] Error while reading message", logger.Field("error", err.Error()))
		return
	}

	event, err := newEvent(m.Value)
	if err != nil {
		logger.E(ctx, err, "[KafkaConsumer] Error while unmarshalling event", logger.Field("event", string(m.Value)), logger.Field("error", err.Error()))
		return
	}

	// Process the Event
	err = c.handler.Handle(ctx, event)

	retries := 0
	for err != nil && retries < c.opts.MaxRetry {
		logger.I(ctx, "retrying")
		backoff := exponentialBackoffWithJitter(retries)
		time.Sleep(backoff)

		// Process the Event
		err = c.handler.Handle(ctx, event)
		retries++
	}

	if err != nil {
		logger.E(ctx, err, "[KafkaConsumer] Processing of event failed",
			logger.Field("event_id", event.ID),
			logger.Field("error", err.Error()))
	}
}

func convertToMap(bytes []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal(bytes, &m)
	return m, err
}

func exponentialBackoffWithJitter(i int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	backoff := minOffset * (1 << i)
	jitter := time.Duration(rand.Int63n(int64(maxJitter/time.Millisecond))) * time.Millisecond
	if backoff < maxBackoff {
		return backoff + jitter
	}
	return maxBackoff + jitter
}

func (c *kafkaConsumer) Close() {
	c.reader.Close()
}
