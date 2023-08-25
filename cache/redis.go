package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"sparrow/logger"
)

type redisClient struct {
	pool *redis.Pool
}

type RedisCacheOpts struct {
	DB                    int
	Host                  string
	Password              string
	MaxIdleConnection     int
	MaxActiveConnection   int
	IdleConnectionTimeout time.Duration
	MaxConnectionLifetime time.Duration
}

var redisCacheClient *redisClient

func NewRedisCache() Cache {
	return redisCacheClient
}

func InitRedisCache(opts *RedisCacheOpts) {
	pool := initRedisPool(opts)

	redisCacheClient = &redisClient{
		pool: pool,
	}

	conn := pool.Get()

	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		panic(fmt.Sprintf("failed to connect redis %v", err.Error()))
	}
}

func initRedisPool(opts *RedisCacheOpts) *redis.Pool {
	return &redis.Pool{
		MaxIdle:         opts.MaxIdleConnection,
		MaxActive:       opts.MaxActiveConnection,
		IdleTimeout:     opts.IdleConnectionTimeout, // Setting timeout so that the workers are not blocked
		MaxConnLifetime: opts.MaxConnectionLifetime,
		Dial: func() (redis.Conn, error) {
			passwordOption := redis.DialPassword(opts.Password)
			dbOption := redis.DialDatabase(opts.DB)
			c, dialErr := redis.Dial("tcp", opts.Host, passwordOption, dbOption)
			if dialErr != nil {
				panic(fmt.Sprintf("dial error: %s", dialErr.Error()))
			}
			return c, dialErr
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// NOTE: Please don't put huge json into the case
// TODO: Add safe guard to checkout the value to be cached, if bigger than threshold should raise exception
func (c *redisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	conn := c.pool.Get()
	defer conn.Close()

	val, _ := json.Marshal(value)

	_, err := redis.String(
		conn.Do(
			"SETEX",
			key,
			int(ttl.Seconds()),
			string(val),
		),
	)

	if err != nil {
		logger.E(ctx, err, "[RedisCache] failed to cache the value",
			logger.Field("key", key),
			logger.Field("value", value),
		)
		return false
	}

	return true
}

func (c *redisClient) Get(ctx context.Context, key string) (interface{}, bool) {
	conn := c.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))

	if err != nil {
		logger.E(ctx, err, "[RedisCache] failed while checking key in cache",
			logger.Field("key", key),
		)
	}

	if !exists {
		return nil, false
	}

	result, err := redis.String(conn.Do("GET", key))
	if err != nil {
		logger.E(ctx, err, "[RedisCache] error while getting value from cache",
			logger.Field("key", key),
		)
		return nil, false
	}
	return result, true
}

func (c *redisClient) GetStruct(ctx context.Context, key string, dest interface{}) bool {
	conn := c.pool.Get()
	defer conn.Close()

	result, found := c.Get(ctx, key)
	if !found {
		return false
	}

	if err := json.Unmarshal([]byte(result.(string)), dest); err != nil {
		logger.E(ctx, err, "[RedisCache] error while unmarshaling value fetched from cache",
			logger.Field("key", key),
			logger.Field("value", result),
		)
		return false
	}
	return true
}

func (c *redisClient) Exists(ctx context.Context, key string) bool {
	conn := c.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		logger.E(ctx, err, "[RedisCache] failed while checking key in cache",
			logger.Field("key", key),
		)
		return false
	}

	return exists
}

func CloseRedisCache() {
	err := redisCacheClient.pool.Close()
	if err != nil {
		panic(fmt.Sprintf("failed to close redis pool %v", err.Error()))
	}
}
