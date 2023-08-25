package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto"
)

type ristrettoClient struct {
	client *ristretto.Cache
}

var ristrettoCacheClient *ristretto.Cache

func NewRistrettoCache() Cache {
	return &ristrettoClient{
		client: ristrettoCacheClient,
	}
}

func InitRistrettoCache(cost int64, counters int64) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: counters,
		MaxCost:     cost,
		BufferItems: 64,
	})

	if err != nil {
		panic(err)
	}

	ristrettoCacheClient = client
}

// NOTE: Please don't put huge json into the case
// TODO: Add safe guard to checkout the value to be cached, if bigger than threshold should raise exception
func (c *ristrettoClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	resp := c.client.SetWithTTL(key, value, 0, ttl)
	c.client.Wait()
	return resp
}

func (c *ristrettoClient) Get(ctx context.Context, key string) (interface{}, bool) {
	return c.client.Get(key)
}

func (c *ristrettoClient) GetStruct(ctx context.Context, key string, dest interface{}) bool {
	value, found := c.client.Get(key)

	if !found {
		return false
	}

	bytes, err := json.Marshal(value)

	if err != nil {
		return false
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return false
	}

	return true
}

func (c *ristrettoClient) Exists(ctx context.Context, key string) bool {
	_, b := c.client.Get(key)
	return b
}

func CloseRistrettoCache() {
	ristrettoCacheClient.Close()
}
