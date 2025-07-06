package configs

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerEnv  string `mapstructure:"SERVER_ENV"`
	ServerPort string `mapstructure:"SERVER_PORT"`

	PostgresDSN          string `mapstructure:"POSTGRES_DSN"`
	PostgresMigrationDSN string `mapstructure:"POSTGRES_MIGRATION_DSN"`

	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisURI      string `mapstructure:"REDIS_URI"`

	AccessTokenPrivateKey string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
	AccessTokenPublicKey  string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY"`
	AccessTokenExpiresIn  time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRES_IN"`

	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY"`
	RefreshTokenExpiresIn  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRES_IN"`
}

func Load(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(&config); err != nil {
		return
	}

	return
}
