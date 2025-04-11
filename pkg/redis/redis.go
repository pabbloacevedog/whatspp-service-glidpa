package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is a wrapper around the Redis client
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client
func NewClient(addr string) *Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Client{
		client: client,
	}
}

// Set sets a key-value pair in Redis with an expiration time
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value from Redis by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Delete deletes a key from Redis
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Ping pings the Redis server
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.client.Close()
}
