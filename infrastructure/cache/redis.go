package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	rdb *redis.Client
}

func NewRedisClient() *Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})

	return &Client{rdb: rdb}
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, data, expiration).Err()
}

func (c *Client) Get(key string, dest interface{}) error {
	ctx := context.Background()
	
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Delete(key string) error {
	ctx := context.Background()
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(key string) (bool, error) {
	ctx := context.Background()
	count, err := c.rdb.Exists(ctx, key).Result()
	return count > 0, err
}

func (c *Client) Close() error {
	return c.rdb.Close()
}