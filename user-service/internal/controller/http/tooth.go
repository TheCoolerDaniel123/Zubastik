package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

type ToothController struct {
	uc        *usecase.UserUseCase
	validator *validator.Validate
}

func NewToothController(uc *usecase.UserUseCase, v *validator.Validate) *ToothController {
	return &ToothController{
		uc:        uc,
		validator: v,
	}
}

func (c *ToothController) GetTeethHandler(w http.ResponseWriter, r *http.Request) {
	teeth, err := c.uc.GetTeeth(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, teeth)
}

func (c *ToothController) UpdateTeethHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.UpdateTeethBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.UpdateTeeth(r.Context(), *body); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}
