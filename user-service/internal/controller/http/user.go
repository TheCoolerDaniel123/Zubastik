package http

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/usecase"
)

var errUserNotFound = errors.New("user not found")

type UserController struct {
	uc        *usecase.UserUseCase
	validator *validator.Validate
}

func NewUserController(uc *usecase.UserUseCase, v *validator.Validate) *UserController {
	return &UserController{
		uc:        uc,
		validator: v,
	}
}

func (c *UserController) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := c.uc.GetUserByUUID(r.Context())
	if err == nil && user == nil {
		err = errUserNotFound
	}
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusNotFound, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, user)
}

func (c *UserController) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.UpdateProfileBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	err = c.uc.UpdateProfile(r.Context(), *body)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}
