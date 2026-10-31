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
		GRPCPort int
		Database Database
		Auth     Auth
		GigaChat GigaChat
	}
	Database struct {
		Host     string
		User     string
		Password string
		Name     string
	}
	Auth struct {
		JWTSecret      string
		AccessTokenTTL time.Duration
	}
	GigaChat struct {
		AuthorizationKey string
		Scope            string
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

	cfg.EnvType = EnvType(viper.GetString("ENV_TYPE"))
	if cfg.EnvType == "" {
		cfg.EnvType = Local
	}

	cfg.RunPort = viper.GetInt("RUN_PORT")
	if cfg.RunPort == 0 {
		cfg.RunPort = 8081
	}

	cfg.GRPCPort = viper.GetInt("GRPC_PORT")
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50051
	}

	cfg.Database.Host = viper.GetString("DATABASE_HOST")
	cfg.Database.User = viper.GetString("DATABASE_USER")
	cfg.Database.Password = viper.GetString("DATABASE_PASSWORD")
	cfg.Database.Name = viper.GetString("DATABASE_NAME")

	cfg.Auth.JWTSecret = viper.GetString("JWT_SECRET")

	cfg.GigaChat.AuthorizationKey = viper.GetString("GIGACHAT_AUTHORIZATION_KEY")
	cfg.GigaChat.Scope = viper.GetString("GIGACHAT_SCOPE")
	if cfg.GigaChat.Scope == "" {
		cfg.GigaChat.Scope = "GIGACHAT_API_PERS"
	}

	var err error
	cfg.Auth.AccessTokenTTL, err = time.ParseDuration(viper.GetString("ACCESS_TOKEN_TTL"))
	if err != nil {
		cfg.Auth.AccessTokenTTL = 15 * time.Minute
	}

	return cfg
}
