package infrastructure

import(
	"github.com/redis/go-redis/v9"
	"lqkhoi-go-http-api/config"
)

func NewRedisConnection(cfg config.RedisConfig) *redis.Client{
	return redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Password, // no password set
        DB:       0,  // use default DB
    })
}

