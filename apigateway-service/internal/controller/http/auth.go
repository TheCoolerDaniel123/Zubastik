package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
)

type AuthController struct {
	AuthServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewAuthController(url string, cli *clients.HTTPClient) *AuthController {
	return &AuthController{AuthServiceURL: url, HTTPClient: cli}
}

func (c *AuthController) SendCodeHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.AuthServiceURL, c.HTTPClient, "/api/v1/auth/send_code")
}

func (c *AuthController) VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.AuthServiceURL, c.HTTPClient, "/api/v1/auth/verify_code")
}

func (c *AuthController) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.AuthServiceURL, c.HTTPClient, "/api/v1/auth/refresh_token")
}
