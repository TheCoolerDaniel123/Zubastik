package service

import (
	"fmt"
	"log"
	"net/http"

	"apigateway-service/internal/config"
	ctrl "apigateway-service/internal/controller/http"
	internalhttp "apigateway-service/internal/http"
	"apigateway-service/pkg/clients"
	sharedjwt "shared/jwt"
)

func MustRun(cfg *config.Config) {
	log.Printf("apigateway-service starting with config %+v", cfg)

	httpCli, err := clients.NewHTTPClient()
	if err != nil {
		log.Fatalf("failed to create HTTP client: %v", err)
	}

	authCtrl := ctrl.NewAuthController(cfg.AuthServiceURL, httpCli)
	userCtrl := ctrl.NewUserController(cfg.UserServiceURL, httpCli)
	calendarCtrl := ctrl.NewCalendarController(cfg.UserServiceURL, httpCli)
	doctorCtrl := ctrl.NewDoctorController(cfg.UserServiceURL, httpCli)
	messagingCtrl := ctrl.NewMessagingController(cfg.UserServiceURL, httpCli)
	chatCtrl := ctrl.NewAiChatController(cfg.UserServiceURL, httpCli)
	articleCtrl := ctrl.NewArticleController(cfg.UserServiceURL, httpCli)

	jwtCfg := &sharedjwt.JWTConfig{JWTSecret: cfg.JWTSecret}

	router := internalhttp.NewRouter(authCtrl, userCtrl, calendarCtrl, doctorCtrl, messagingCtrl, chatCtrl, articleCtrl, jwtCfg)

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", cfg.RunPort),
	}

	log.Printf("apigateway-service listening on port %d", cfg.RunPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
