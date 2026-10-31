package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
)

type MessagingController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewMessagingController(url string, cli *clients.HTTPClient) *MessagingController {
	return &MessagingController{UserServiceURL: url, HTTPClient: cli}
}

func (c *MessagingController) StartPatientConversationHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *MessagingController) GetOrCreateConversationHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *MessagingController) GetDoctorConversationsHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *MessagingController) GetPatientConversationsHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *MessagingController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}

func (c *MessagingController) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, r.URL.RequestURI())
}
