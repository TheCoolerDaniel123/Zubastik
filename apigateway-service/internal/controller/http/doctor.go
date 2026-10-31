package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
)

type DoctorController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewDoctorController(url string, cli *clients.HTTPClient) *DoctorController {
	return &DoctorController{UserServiceURL: url, HTTPClient: cli}
}

func (c *DoctorController) GetPatientProfileHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) GetPatientTeethHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) CreateToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) GetToothNotesHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) UpdateToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) DeleteToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) CreateRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *DoctorController) GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}
