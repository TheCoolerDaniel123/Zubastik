package service

import (
	"auth-service/internal/config"
	controllerhttp "auth-service/internal/controller/http"
	internalhttp "auth-service/internal/http"
	redisrepo "auth-service/internal/storage/cache/redis"
	userrepo "auth-service/internal/storage/database/postgres"
	"auth-service/internal/usecase"
	grpcclient "auth-service/pkg/clients/grpc"
	redisclient "auth-service/pkg/clients/redis"
	smsclient "auth-service/pkg/clients/sms"
	pgclient "shared/clients/postgres"

	"fmt"
	"log"
	"net/http"
)

func MustRun(cfg *config.Config) {
	log.Printf("auth-service starting with config %+v", cfg)

	pg, err := pgclient.NewPostgresClient(&pgclient.Database{
		Host:     cfg.Database.Host,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
	})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	rd := redisclient.NewRedisClient(&cfg.Cache)

	grpcCli, err := grpcclient.NewGRPCClient(cfg.GRPC)
	if err != nil {
		log.Fatalf("failed to connect to grpc: %v", err)
	}

	sms := smsclient.NewMockSMSSender()

	ur := userrepo.NewUserRepo(pg)
	cr := redisrepo.NewCodeRepo(rd)

	uc := usecase.NewAuthUseCase(grpcCli, cr, ur, sms, &cfg.Auth)
	authCtrl := controllerhttp.NewAuthController(uc)

	router := internalhttp.NewRouter(authCtrl)

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", cfg.RunPort),
	}

	log.Printf("auth-service starting on port %d", cfg.RunPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
