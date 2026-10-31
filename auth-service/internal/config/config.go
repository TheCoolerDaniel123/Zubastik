package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type EnvType string

const (
	Local EnvType = "local"
	Dev   EnvType = "dev"
	Prod  EnvType = "prod"
)

type (
	Config struct {
		EnvType  EnvType
		RunPort  int
		Database Database
		Cache    Cache
		Auth     Auth
		GRPC     string
	}
	Database struct {
		Host     string
		User     string
		Password string
		Name     string
	}
	Cache struct {
		Host     string
		Port     string
		Password string
	}
	Auth struct {
		JWTSecret           string
		AccessTokenTTL      time.Duration
		RefreshTokenTTL     time.Duration
		VerificationCodeTTL time.Duration
	}
)

func MustLoad() *Config {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, using environment variables only")
	}

	cfg := &Config{}

	// ENV_TYPE
	cfg.EnvType = EnvType(viper.GetString("ENV_TYPE"))
	if cfg.EnvType == "" {
		cfg.EnvType = Local
	}

	// RUN_PORT
	cfg.RunPort = viper.GetInt("RUN_PORT")
	if cfg.RunPort == 0 {
		cfg.RunPort = 8080
	}

	cfg.GRPC = viper.GetString("GRPC")
	if cfg.GRPC == "" {
		cfg.GRPC = "user-service:50051"
	}

	// Database
	cfg.Database.Host = viper.GetString("DATABASE_HOST")
	cfg.Database.User = viper.GetString("DATABASE_USER")
	cfg.Database.Password = viper.GetString("DATABASE_PASSWORD")
	cfg.Database.Name = viper.GetString("DATABASE_NAME")

	// Cache
	cfg.Cache.Host = viper.GetString("CACHE_HOST")
	cfg.Cache.Port = viper.GetString("CACHE_PORT")
	cfg.Cache.Password = viper.GetString("CACHE_PASSWORD")

	// Auth
	cfg.Auth.JWTSecret = viper.GetString("JWT_SECRET")

	var err error
	cfg.Auth.AccessTokenTTL, err = time.ParseDuration(viper.GetString("ACCESS_TOKEN_TTL"))
	if err != nil {
		cfg.Auth.AccessTokenTTL = 15 * time.Minute
	}

	cfg.Auth.RefreshTokenTTL, err = time.ParseDuration(viper.GetString("REFRESH_TOKEN_TTL"))
	if err != nil {
		cfg.Auth.RefreshTokenTTL = 24 * time.Hour
	}

	cfg.Auth.VerificationCodeTTL, err = time.ParseDuration(viper.GetString("VERIFICATION_CODE_TTL"))
	if err != nil {
		cfg.Auth.VerificationCodeTTL = 5 * time.Minute
	}

	return cfg
}
