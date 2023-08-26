package worker

import (
	"context"
)

type middleware struct {
}

type Middleware interface {
	ProcessTask(context.Context, *Task) error
}
