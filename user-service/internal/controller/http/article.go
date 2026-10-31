package http

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

type ArticleController struct {
	uc        *usecase.ArticleUseCase
	validator *validator.Validate
}

func NewArticleController(uc *usecase.ArticleUseCase, v *validator.Validate) *ArticleController {
	return &ArticleController{uc: uc, validator: v}
}

func (c *ArticleController) GetApprovedArticlesHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := c.uc.GetApprovedArticles(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, articles)
}

func (c *ArticleController) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.CreateArticleBody](r, c.validator)
	if err != nil {
		log.Printf("article create: bind error: %v", err)
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	article, err := c.uc.CreateArticle(r.Context(), *body)
	if err != nil {
		log.Printf("article create: usecase error: %v", err)
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, article)
}

func (c *ArticleController) GetAllArticlesHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := c.uc.GetAllArticles(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, articles)
}

func (c *ArticleController) ApproveArticleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "article_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.ApproveArticle(r.Context(), id); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}

func (c *ArticleController) RejectArticleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "article_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.RejectArticle(r.Context(), id); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}
