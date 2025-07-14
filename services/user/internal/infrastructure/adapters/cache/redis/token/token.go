package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/adapters/cache/redis"
	"github.com/go-redis/redis/v8"
)

type TokenAdapter struct {
	redisClient *redis.Client
	logger      logger.Logger
	cfg         *configs.Config
}

func NewTokenAdapter(log logger.Logger, cfg *configs.Config) (domain.TokenAdapter, error) {
	loggerTag := "adapters.cache.redis.token.newTokenAdapter"

	log.Info(loggerTag, "Initializing the token adapter")

	log.Info(loggerTag, "Initializing redis client")
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisURI,
			Password: cfg.RedisPassword,
		},
	)

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Error(loggerTag, ErrConnecting, logger.Field{
			Key:   "error",
			Value: err.Error(),
		})

		return nil, fmt.Errorf("%s: %v", ErrConnecting, err)
	}
	log.Info(loggerTag, "Connection to the redis has been completed")

	return &TokenAdapter{
		redisClient,
		log,
		cfg,
	}, nil
}

func (ta *TokenAdapter) Set(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error {
	return ta.redisClient.Set(ctx, fmt.Sprintf("refresh_token:%s", userID), refreshToken, expiresIn).Err()
}

func (ta *TokenAdapter) Get(ctx context.Context, userID string) (string, error) {
	return ta.redisClient.Get(ctx, fmt.Sprintf("refresh_token:%s", userID)).Result()
}

func (ta *TokenAdapter) Del(ctx context.Context, userID string) error {
	return ta.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", userID)).Err()
}
