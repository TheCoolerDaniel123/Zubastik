package config

import (
	"log"

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
		EnvType        EnvType
		RunPort        int
		AuthServiceURL string
		UserServiceURL string
		JWTSecret      string
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

	cfg.AuthServiceURL = viper.GetString("AUTH_SERVICE_URL")
	if cfg.AuthServiceURL == "" {
		log.Fatal("AUTH_SERVICE_URL is required (e.g. http://auth-service:8080)")
	}

	cfg.UserServiceURL = viper.GetString("USER_SERVICE_URL")
	if cfg.UserServiceURL == "" {
		log.Fatal("USER_SERVICE_URL is required (e.g. http://user-service:8081)")
	}

	cfg.JWTSecret = viper.GetString("JWT_SECRET")

	return cfg
}
