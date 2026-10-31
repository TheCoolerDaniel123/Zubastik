package main

import (
	"auth-service/internal/config"
	"auth-service/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	config := config.MustLoad()
	service.MustRun(config)

	// db, _ := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	// userRepo := postgres.NewUserRepo(db)
	// codeRepo := memory.NewAuthRepo()
	// smsSender := sms.NewMockSMSSender()
	// jwtCfg := &jwt.JWTConfig{
	// 	JWTSecret:       "sverh_secret",
	// 	AccessTokenTTL:  15 * time.Minute,
	// 	RefreshTokenTTL: 7 * 24 * time.Hour}

	// uc := usecase.NewAuthUseCase(codeRepo, userRepo, smsSender, jwtCfg)

	// ctrl := http_ctr.NewAuthController(uc)
	// ctrl_prf := http_ctr.NewProfileController(userRepo)

	// http.HandleFunc("/send_code", ctrl.SendCodeHandler)
	// http.HandleFunc("/verify_code", ctrl.VerifyCodeHandler)
	// http.HandleFunc("/refresh", ctrl.RefreshTokenHandler)
	// http.Handle("/profile", http_ctr.JWTMiddleware(jwtCfg)(http.HandlerFunc(ctrl_prf.GetProfileHandler)))

	// log.Println("Auth service running at :8080")
	// http.ListenAndServe(":8080", nil)
}
