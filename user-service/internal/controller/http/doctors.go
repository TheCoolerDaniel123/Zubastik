package http

import (
	"net/http"

	"shared/entity"
	sharedhttp "shared/http"
	"user-service/internal/usecase"
)

type DoctorsController struct {
	uc *usecase.DoctorsUseCase
}

func NewDoctorsController(uc *usecase.DoctorsUseCase) *DoctorsController {
	return &DoctorsController{uc: uc}
}

func (c *DoctorsController) GetDoctorsHandler(w http.ResponseWriter, r *http.Request) {
	profiles, err := c.uc.GetAll(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	if profiles == nil {
		profiles = []entity.DoctorProfile{}
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, profiles)
}
