package infrastructure

import(
    "lqkhoi-go-http-api/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(cfg config.RedisConfig) *redis.Client{
	return redis.NewClient(&redis.Options{
        Addr:     cfg.Addr(),
        Password: cfg.Password, // no password set
        DB:       0,  // use default DB
    })
}

