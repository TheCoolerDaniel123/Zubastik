package http

import "net/http"

type AuthController interface {
	SendCodeHandler(w http.ResponseWriter, r *http.Request)
	VerifyCodeHandler(w http.ResponseWriter, r *http.Request)
	RefreshTokenHandler(w http.ResponseWriter, r *http.Request)
}
