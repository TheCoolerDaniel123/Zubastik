package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type (
	errorResponseBody struct {
		Message string `json:"message" example:"error description"`
	}
)

func BindAndValidate[T any](r *http.Request, validate *validator.Validate) (*T, error) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := validate.StructCtx(r.Context(), &body); err != nil {
		return nil, err
	}
	return &body, nil
}

func ResponseWithSuccess(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
	w.WriteHeader(statusCode)
	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}

func ResponseWithError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	w.WriteHeader(statusCode)

	if ew, ok := w.(interface{ SetError(error) }); ok {
		ew.SetError(err)
	}

	if statusCode >= 500 {
		json.NewEncoder(w).Encode(errorResponseBody{Message: "internal server error"})
	} else {
		json.NewEncoder(w).Encode(errorResponseBody{Message: err.Error()})
	}
}
