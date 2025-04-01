package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/KinMod-ui/thelaLocator/proto"
	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	GetUser(ctx context.Context, key string) (*proto.User, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	SetUser(ctx context.Context, key string, user *proto.User, expiration time.Duration) error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string, password string, db int) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}
	return &redisClient{client: client}, nil
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) GetUser(ctx context.Context, key string) (*proto.User, error) {
	var user proto.User
	err := r.client.Get(ctx, key).Scan(&user)
	if err != nil {
		return &proto.User{}, err
	} else {
		return &user, nil
	}
}

func (r *redisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisClient) SetUser(ctx context.Context, key string, user *proto.User, expiration time.Duration) error {
	return r.client.Set(ctx, key, user, expiration).Err()
}
