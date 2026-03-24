package worker

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var (
	enqueuerInstance *enqueuer
	once             sync.Once
)

type enqueuer struct {
	client *asynq.Client
}

type EnqueuerOpts struct {
	PoolSize int
	RedisUrl string
	RedisDB int
}

type Enqueuer interface {
	EnqueueUniqueTask(*Task, uniqueTTL time.Duration) error
	EnqueueUniqueTaskIn(*Task, time.Duration, uniqueTTL time.Duration) error
}

func NewEnqueuer(opts *EnqueuerOpts) Enqueuer {
	once.Do(func() {
		redisConnection := asynq.RedisClientOpt{
			PoolSize: opts.PoolSize,
			Addr:     opts.RedisUrl,
			DB: opts.RedisDB,
		}

		enqueuerInstance = &enqueuer{
			client: asynq.NewClient(redisConnection),
		}
	})

	return enqueuerInstance
}

func (e *enqueuer) EnqueueUniqueTask(task *Task, uniqueTTL time.Duration) error {
	taskID := uuid.New().String()
	bytes, err := json.Marshal(task.Payload)
	if err != nil {
		return err
	}

	asynqTask := asynq.NewTask(task.Name, bytes)
	opts := []asynq.Option{asynq.TaskID(taskID), asynq.Unique(uniqueTTL), asynq.MaxRetry(task.Retry), asynq.Timeout(task.Timeout)}
	_, err = e.client.Enqueue(
		asynqTask,
		opts...,
	)
	return err
}

func (e *enqueuer) EnqueueUniqueTaskIn(task *Task, delay time.Duration, uniqueTTL time.Duration) error {
	taskID := uuid.New().String()
	bytes, err := json.Marshal(task.Payload)
	if err != nil {
		return err
	}

	asynqTask := asynq.NewTask(task.Name, bytes)
	opts := []asynq.Option{asynq.TaskID(taskID), asynq.Unique(uniqueTTL), asynq.MaxRetry(task.Retry), asynq.Timeout(task.Timeout), asynq.ProcessIn(delay)}
	_, err = e.client.Enqueue(
		asynqTask,
		opts...,
	)
	return err
}

func (e *enqueuer) Close() error {
	return e.client.Close()
}
