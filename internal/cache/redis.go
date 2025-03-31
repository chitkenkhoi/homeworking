package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"lqkhoi-go-http-api/pkg/structs"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, exp int) error
	Del(ctx context.Context, key string) error
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) CacheRepository {
	return &redisRepository{
		client: client,
	}
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	cmd := r.client.Get(ctx, key)
	val, err := cmd.Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			slog.Error("key does not exist", "key", key)
			return "", structs.ErrRedisKeyNotExist
		}
		slog.Error("redis Get failed", "key", key, "error", err)
		return "", structs.ErrRedisConnection
	}
	slog.Debug("Successfully get key", "key", key, "value", val)
	return val, nil
}

func (r *redisRepository) Set(ctx context.Context, key string, value any, exp int) error {
	cmd := r.client.Set(ctx, key, value, time.Duration(exp)*time.Minute)
	_, err := cmd.Result()

	if err != nil {
		slog.Error("redis Set failed", "key", key, "value", value, "error", err)
		return structs.ErrRedisConnection
	}
	slog.Debug("Successfully set key with value", "key", key, "value", value, "exp", exp)
	return nil
}

func (r *redisRepository) Del(ctx context.Context, key string) error {
	cmd := r.client.Del(ctx, key)
	deletedCount, err := cmd.Result()

	if err != nil {
		slog.Error("redis Del failed", "key", key, "error", err)
		return structs.ErrRedisConnection
	} else if deletedCount == 0 {
		slog.Error("key does not exist", "key", key)
		return structs.ErrRedisKeyNotExist
	}
	slog.Debug("Successfully delete key", "key", key)
	return nil
}

func (r *redisRepository) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *redisRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
    if err != nil {
        if errors.Is(err, redis.Nil) {
             return 0, err
        }
        return 0, err
    }
	return ttl, nil
}