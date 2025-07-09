package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	"github.com/go-redis/redis/v8"
)

type TokenRepository struct {
	redisClient *redis.Client
	logger      logger.Logger
	cfg         *configs.Config
}

func NewTokenRepository(repoLogger logger.Logger, cfg *configs.Config) (domain.TokenRepository, error) {
	loggerTag := "token.repository.newTokenRepository"

	repoLogger.Info(loggerTag, "Initializing the token repository")

	repoLogger.Info(loggerTag, "Initializing redis client")
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisURI,
			Password: cfg.RedisPassword,
		},
	)

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		repoLogger.Error(loggerTag, ErrConnecting, logger.Field{
			Key:   "error",
			Value: err.Error(),
		})

		return nil, fmt.Errorf("%s: %v", ErrConnecting, err)
	}
	repoLogger.Info(loggerTag, "Connection to the redis has been completed")

	return &TokenRepository{
		redisClient,
		repoLogger,
		cfg,
	}, nil
}

func (r *TokenRepository) Set(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error {
	loggerTag := "token.repository.set"

	if err := r.redisClient.Set(ctx, fmt.Sprintf("refresh_token:%s", userID), refreshToken, expiresIn).Err(); err != nil {
		r.logger.Error(loggerTag, fmt.Sprintf("failed set token in redis: %v", err))

		return err
	}

	return nil
}

func (r *TokenRepository) Get(ctx context.Context, userID string) (string, error) {
	loggerTag := "token.repository.get"

	refreshToken, err := r.redisClient.Get(ctx, fmt.Sprintf("refresh_token:%s", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenNotFound
		}

		r.logger.Error(loggerTag, fmt.Sprintf("failed get token from redis: %v", err))

		return "", err
	}

	return refreshToken, nil
}

func (r *TokenRepository) Del(ctx context.Context, userID string) error {
	loggerTag := "token.repository.del"

	if err := r.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", userID)).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrTokenNotFound
		}

		r.logger.Error(loggerTag, fmt.Sprintf("failed del token from redis: %v", err))

		return err
	}

	return nil
}
