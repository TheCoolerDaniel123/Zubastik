package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"shared/entity"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

type DoctorController struct {
	uc        *usecase.DoctorUseCase
	validator *validator.Validate
}

func NewDoctorController(uc *usecase.DoctorUseCase, v *validator.Validate) *DoctorController {
	return &DoctorController{uc: uc, validator: v}
}

func (c *DoctorController) GetPatientProfileHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	user, err := c.uc.GetPatientByUUID(r.Context(), patientUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusNotFound, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, user)
}

func (c *DoctorController) GetPatientTeethHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	teeth, err := c.uc.GetPatientTeeth(r.Context(), patientUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, teeth)
}

func (c *DoctorController) CreateToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	body, err := sharedhttp.BindAndValidate[dto.CreateToothNoteBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	note, err := c.uc.CreateToothNote(r.Context(), entity.ToothNote{
		PatientUUID: patientUUID,
		ToothIndex:  body.ToothIndex,
		Note:        body.Note,
	})
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, note)
}

func (c *DoctorController) GetToothNotesHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	notes, err := c.uc.GetToothNotes(r.Context(), patientUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, notes)
}

func (c *DoctorController) UpdateToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	noteID, err := uuid.Parse(chi.URLParam(r, "note_id"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	body, err := sharedhttp.BindAndValidate[dto.UpdateToothNoteBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.UpdateToothNote(r.Context(), noteID, body.Note); err != nil {
		sharedhttp.ResponseWithError(w, r, noteStatusCode(err), err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}

func (c *DoctorController) DeleteToothNoteHandler(w http.ResponseWriter, r *http.Request) {
	noteID, err := uuid.Parse(chi.URLParam(r, "note_id"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.DeleteToothNote(r.Context(), noteID); err != nil {
		sharedhttp.ResponseWithError(w, r, noteStatusCode(err), err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}

func (c *DoctorController) CreateRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	body, err := sharedhttp.BindAndValidate[dto.CreateRecommendationBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	rec, err := c.uc.CreateRecommendation(r.Context(), entity.Recommendation{
		PatientUUID: patientUUID,
		Text:        body.Text,
	})
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, rec)
}

func (c *DoctorController) GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	recs, err := c.uc.GetRecommendations(r.Context(), patientUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, recs)
}

// noteStatusCode maps tooth-note errors to HTTP status codes.
func noteStatusCode(err error) int {
	if errors.Is(err, database.ErrNoteNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
