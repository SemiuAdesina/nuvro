package redis

import (
	"context"
	"errors"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type Client struct{ client *redisclient.Client }

func Open(url string) (*Client, error) {
	options, err := redisclient.ParseURL(url)
	if err != nil {
		return nil, errors.New("redis: invalid configuration")
	}
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = time.Second
	options.WriteTimeout = time.Second
	options.ContextTimeoutEnabled = true
	options.MaxRetries = -1
	return &Client{client: redisclient.NewClient(options)}, nil
}

func (*Client) Name() string                      { return "redis" }
func (c *Client) Check(ctx context.Context) error { return c.client.Ping(ctx).Err() }
func (c *Client) Close() error                    { return c.client.Close() }
