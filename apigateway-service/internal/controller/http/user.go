package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
)

type UserController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewUserController(url string, cli *clients.HTTPClient) *UserController {
	return &UserController{UserServiceURL: url, HTTPClient: cli}
}

func (c *UserController) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/")
}

func (c *UserController) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/")
}

func (c *UserController) GetTeethHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/teeth")
}

func (c *UserController) UpdateTeethHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/teeth")
}

func (c *UserController) GetBrushScheduleHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/brush-schedule")
}

func (c *UserController) UpdateBrushScheduleHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/profile/brush-schedule")
}

func (c *UserController) GetDoctorsHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/doctors/")
}
