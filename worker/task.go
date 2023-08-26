package worker

import (
	"time"
)

type Task struct {
	Name    string
	Retry   int
	Timeout time.Duration
	Payload interface{}
}
