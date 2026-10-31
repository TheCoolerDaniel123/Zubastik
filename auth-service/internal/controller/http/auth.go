package http

import (
	"auth-service/api/dto"
	"auth-service/internal/usecase"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	sharedhttp "shared/http"
)

type AuthController struct {
	uc        *usecase.AuthUseCase
	validator *validator.Validate
}

func NewAuthController(uc *usecase.AuthUseCase) *AuthController {
	return &AuthController{uc: uc, validator: validator.New()}
}

func (c *AuthController) SendCodeHandler(w http.ResponseWriter, r *http.Request) {
	var body struct{ Phone string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := c.uc.SendAuthorizationCode(r.Context(), body.Phone); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}

func (c *AuthController) VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.ConfirmCodeBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	tokenPair, isNewUser, err := c.uc.VerifyCode(r.Context(), body.Phone, body.Code)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCode:
			sharedhttp.ResponseWithError(w, r, http.StatusUnauthorized, err)
		default:
			sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, dto.VerifyCodeResponse{
		TokenPair: tokenPair,
		IsNewUser: isNewUser,
	})
}

func (c *AuthController) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	tokenPair, err := c.uc.RefreshToken(r.Context(), body.AccessToken, body.RefreshToken)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusUnauthorized, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, tokenPair)
}
