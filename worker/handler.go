package worker

import (
	"github.com/hibiken/asynq"
)

type Handler struct {
	Name        string
	HandlerFunc asynq.HandlerFunc
}
