package consumer

import (
	"context"
)

type EventHandler interface {
	Name() string
	Handle(ctx context.Context, event *Event) error
}
