package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
	"github.com/go-chi/chi"
)

type ArticleController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewArticleController(url string, cli *clients.HTTPClient) *ArticleController {
	return &ArticleController{UserServiceURL: url, HTTPClient: cli}
}

func (c *ArticleController) GetApprovedArticlesHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/lectures/")
}

func (c *ArticleController) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/doctor/lectures/")
}

func (c *ArticleController) GetAllArticlesHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/admin/lectures/")
}

func (c *ArticleController) ApproveArticleHandler(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "article_id")
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/admin/lectures/"+articleID+"/approve")
}

func (c *ArticleController) RejectArticleHandler(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "article_id")
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/admin/lectures/"+articleID+"/reject")
}
