package cache

import (
	"context"
	"fmt"
	"time"
)

const (
	serviceNamespace = "sparrow"
)

type Cache interface {
	Set(context.Context, string, interface{}, time.Duration) bool
	Get(context.Context, string) (interface{}, bool)
	GetStruct(context.Context, string, interface{}) bool
	Exists(context.Context, string) bool
}

func GetKey(slugs ...string) string {
	finalKey := serviceNamespace

	for _, slug := range slugs {
		finalKey = fmt.Sprintf("%s:%s", finalKey, slug)
	}

	return finalKey
}
