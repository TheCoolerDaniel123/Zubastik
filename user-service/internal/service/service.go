package service

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	pb "shared/api/pb"
	pgclient "shared/clients/postgres"
	sharedjwt "shared/jwt"
	"user-service/internal/config"
	controllergrpc "user-service/internal/controller/grpc"
	controllerhttp "user-service/internal/controller/http"
	internalhttp "user-service/internal/http"
	"user-service/internal/storage/database/postgres"
	"user-service/internal/usecase"
	aiclient "user-service/pkg/clients/ai"
)

func MustRun(cfg *config.Config) {
	log.Printf("user-service starting with config %+v", cfg)

	pg, err := pgclient.NewPostgresClient(&pgclient.Database{
		Host:     cfg.Database.Host,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
	})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	userRepo := postgres.NewUserRepo(pg)
	toothRepo := postgres.NewToothRepo(pg)
	brushRepo := postgres.NewBrushingTimesRepo(pg)
	calendarRepo := postgres.NewCalendarRepo(pg)
	noteRepo := postgres.NewToothNoteRepo(pg)
	recRepo := postgres.NewRecommendationRepo(pg)
	chatRepo := postgres.NewChatRepo(pg)
	convRepo := postgres.NewConversationRepo(pg)
	dmRepo := postgres.NewDirectMessageRepo(pg)
	articleRepo := postgres.NewArticleRepo(pg)
	doctorProfileRepo := postgres.NewDoctorProfileRepo(pg)

	uc := usecase.NewUserUseCase(userRepo, toothRepo, brushRepo, calendarRepo)
	doctorUC := usecase.NewDoctorUseCase(userRepo, toothRepo, noteRepo, recRepo)
	articleUC := usecase.NewArticleUseCase(articleRepo)
	var aiProvider usecase.AIProvider
	if cfg.GigaChat.AuthorizationKey != "" {
		aiProvider = aiclient.NewGigaChatProvider(cfg.GigaChat.AuthorizationKey, cfg.GigaChat.Scope)
	} else {
		aiProvider = aiclient.NewStubAIProvider()
	}
	chatUC := usecase.NewAiChatUseCase(chatRepo, aiProvider)
	messagingUC := usecase.NewMessagingUseCase(convRepo, dmRepo)
	doctorsUC := usecase.NewDoctorsUseCase(doctorProfileRepo)

	v := validator.New()
	userCtrl := controllerhttp.NewUserController(uc, v)
	toothCtrl := controllerhttp.NewToothController(uc, v)
	brushCtrl := controllerhttp.NewBrushController(uc, v)
	calendarCtrl := controllerhttp.NewCalendarController(uc, v)
	doctorCtrl := controllerhttp.NewDoctorController(doctorUC, v)
	messagingCtrl := controllerhttp.NewMessagingController(messagingUC, v)
	chatCtrl := controllerhttp.NewAiChatController(chatUC, v)
	articleCtrl := controllerhttp.NewArticleController(articleUC, v)
	doctorsListCtrl := controllerhttp.NewDoctorsController(doctorsUC)

	jwtCfg := &sharedjwt.JWTConfig{
		JWTSecret:      cfg.Auth.JWTSecret,
		AccessTokenTTL: cfg.Auth.AccessTokenTTL,
	}
	router := internalhttp.NewRouter(userCtrl, toothCtrl, brushCtrl, calendarCtrl, doctorCtrl, messagingCtrl, chatCtrl, articleCtrl, doctorsListCtrl, jwtCfg)

	grpcCtrl := controllergrpc.NewUserGRPCController(uc)
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, grpcCtrl)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen gRPC: %v", err)
	}

	go func() {
		log.Printf("user-service gRPC listening on port %d", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", cfg.RunPort),
	}

	log.Printf("user-service HTTP listening on port %d", cfg.RunPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
