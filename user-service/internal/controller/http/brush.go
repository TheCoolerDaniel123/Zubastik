package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

type BrushController struct {
	uc        *usecase.UserUseCase
	validator *validator.Validate
}

func NewBrushController(uc *usecase.UserUseCase, v *validator.Validate) *BrushController {
	return &BrushController{
		uc:        uc,
		validator: v,
	}
}

func (c *BrushController) GetBrushScheduleHandler(w http.ResponseWriter, r *http.Request) {
	times, err := c.uc.GetBrushSchedule(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, times)
}

func (c *BrushController) UpdateBrushScheduleHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.UpdateBrushScheduleBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.UpdateBrushSchedule(r.Context(), *body); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}
