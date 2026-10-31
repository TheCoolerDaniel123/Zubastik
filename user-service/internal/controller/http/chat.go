package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

type AiChatController struct {
	uc        *usecase.AiChatUseCase
	validator *validator.Validate
}

func NewAiChatController(uc *usecase.AiChatUseCase, v *validator.Validate) *AiChatController {
	return &AiChatController{uc: uc, validator: v}
}

func (c *AiChatController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.SendMessageBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	reply, err := c.uc.SendMessage(r.Context(), *body)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, reply)
}

func (c *AiChatController) GreetHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.GreetBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	reply, err := c.uc.Greet(r.Context(), *body)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, reply)
}

func (c *AiChatController) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history, err := c.uc.GetHistory(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, history)
}
