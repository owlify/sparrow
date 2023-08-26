package worker

import (
	"github.com/hibiken/asynq"
)

type Handler struct {
	TaskName    string
	HandlerFunc asynq.HandlerFunc
}
