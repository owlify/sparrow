package worker

import (
	"context"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

type WorkerOpts struct {
	PoolSize        int
	RedisUrl        string
	Concurrency     int
	Queues          []*Queue
	ShutdownTimeout time.Duration
}

type worker struct {
	server   *asynq.Server
	handlers []*Handler
}

type Worker interface {
	Start(context.Context) error
	RegisterHandlers([]*Handler)
	Stop()
}

var (
	workerOnce     sync.Once
	workerInstance worker
)

func NewWorker(opts *WorkerOpts) Worker {
	redisClientOpts := asynq.RedisClientOpt{
		PoolSize: opts.PoolSize,
		Addr:     opts.RedisUrl,
	}

	queues := map[string]int{}
	for _, q := range opts.Queues {
		queues[q.Name] = q.Priority
	}

	// Create and configuring Asynq worker server.
	workerServer := asynq.NewServer(redisClientOpts, asynq.Config{
		Concurrency:     opts.Concurrency,
		Queues:          queues,
		ShutdownTimeout: opts.ShutdownTimeout,
	})

	return &worker{
		server: workerServer,
	}
}

func (w *worker) RegisterHandlers(handlers []*Handler) {
	w.handlers = handlers
}

func (w *worker) Start(ctx context.Context) error {
	mux := asynq.NewServeMux()

	for _, handler := range w.handlers {
		mux.HandleFunc(
			handler.TaskName,
			handler.HandlerFunc,
		)
	}

	if err := w.server.Run(mux); err != nil {
		return err
	}

	return nil
}

func (w *worker) Stop() {
	w.server.Stop()
	w.server.Shutdown()
}
