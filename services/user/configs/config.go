package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort int

	PostgresDSN          string
	PostgresMigrationDSN string

	RedisPassword string
	RedisURI      string

	AccessTokenPrivateKey string
	AccessTokenPublicKey  string
	AccessTokenExpiresIn  time.Duration

	RefreshTokenPrivateKey string
	RefreshTokenPublicKey  string
	RefreshTokenExpiresIn  time.Duration
}

func Load() (cfg Config) {
	cfg.ServerPort, _ = strconv.Atoi(os.Getenv("SERVER_PORT"))

	cfg.PostgresDSN = os.Getenv("POSTGRES_DSN")
	cfg.PostgresMigrationDSN = os.Getenv("POSTGRES_MIGRATION_DSN")

	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")
	cfg.RedisURI = os.Getenv("REDIS_URI")

	cfg.AccessTokenPrivateKey = os.Getenv("ACCESS_TOKEN_PRIVATE_KEY")
	cfg.AccessTokenPublicKey = os.Getenv("ACCESS_TOKEN_PUBLIC_KEY")
	cfg.AccessTokenExpiresIn, _ = time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRES_IN"))

	cfg.RefreshTokenPrivateKey = os.Getenv("REFRESH_TOKEN_PRIVATE_KEY")
	cfg.RefreshTokenPublicKey = os.Getenv("REFRESH_TOKEN_PUBLIC_KEY")
	cfg.RefreshTokenExpiresIn, _ = time.ParseDuration(os.Getenv("REFRESH_TOKEN_EXPIRES_IN"))

	return cfg
}
