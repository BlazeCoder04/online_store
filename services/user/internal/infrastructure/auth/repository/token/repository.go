package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/auth/repository"
	"github.com/go-redis/redis/v8"
)

type TokenRepository struct {
	redisClient *redis.Client
	logger      logger.Logger
	cfg         *configs.Config
}

func NewTokenRepository(repoLogger logger.Logger, cfg *configs.Config) (domain.TokenRepository, error) {
	repoLogger.Info("Initializing the token repository")

	repoLogger.Info("Initializing redis client")
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisURI,
			Password: cfg.RedisPassword,
		},
	)

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		repoLogger.Error(ErrConnecting, logger.Field{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, fmt.Errorf("%s: %w", ErrConnecting, err)
	}
	repoLogger.OK("Connection to the redis has been completed")

	return &TokenRepository{
		redisClient,
		repoLogger,
		cfg,
	}, nil
}

func (r *TokenRepository) Set(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error {
	return r.redisClient.Set(ctx, fmt.Sprintf("refresh_token:%s", userID), refreshToken, expiresIn).Err()
}

func (r *TokenRepository) Get(ctx context.Context, userID string) (string, error) {
	return r.redisClient.Get(ctx, fmt.Sprintf("refresh_token:%s", userID)).Result()
}

func (r *TokenRepository) Delete(ctx context.Context, userID string) error {
	return r.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", userID)).Err()
}
