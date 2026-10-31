package http

import (
	"net/http"

	"github.com/go-chi/chi"
)

func NewRouter(authController AuthController) http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1/auth", func(subr chi.Router) {
		subr.Post("/send_code", authController.SendCodeHandler)
		subr.Post("/verify_code", authController.VerifyCodeHandler)
		subr.Post("/refresh_token", authController.RefreshTokenHandler)
	})
	return r
}
