package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
)

type AiChatController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewAiChatController(url string, cli *clients.HTTPClient) *AiChatController {
	return &AiChatController{UserServiceURL: url, HTTPClient: cli}
}

func (c *AiChatController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/ai-chat/message")
}

func (c *AiChatController) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/ai-chat/history")
}

func (c *AiChatController) GreetHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/ai-chat/greet")
}
