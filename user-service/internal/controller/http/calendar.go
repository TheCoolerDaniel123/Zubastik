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

type CalendarController struct {
	uc        *usecase.UserUseCase
	validator *validator.Validate
}

func NewCalendarController(uc *usecase.UserUseCase, v *validator.Validate) *CalendarController {
	return &CalendarController{
		uc:        uc,
		validator: v,
	}
}

func (c *CalendarController) GetCalendarEventsHandler(w http.ResponseWriter, r *http.Request) {
	events, err := c.uc.GetCalendarEvents(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, events)
}

func (c *CalendarController) CreateCalendarEventHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.CreateCalendarEventBody](r, c.validator)
	if err != nil {
		log.Printf("calendar create: bind error: %v", err)
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	log.Printf("calendar create: body=%+v", body)
	event, err := c.uc.CreateCalendarEvent(r.Context(), *body)
	if err != nil {
		log.Printf("calendar create: usecase error: %v", err)
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, event)
}

func (c *CalendarController) DeleteCalendarEventHandler(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "event_id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := c.uc.DeleteCalendarEvent(r.Context(), eventID); err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, nil)
}
