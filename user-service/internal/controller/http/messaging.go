package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	sharedhttp "shared/http"
	"user-service/api/dto"
	"user-service/internal/storage/database"
	"user-service/internal/usecase"
)

type MessagingController struct {
	uc        *usecase.MessagingUseCase
	validator *validator.Validate
}

func NewMessagingController(uc *usecase.MessagingUseCase, v *validator.Validate) *MessagingController {
	return &MessagingController{uc: uc, validator: v}
}

func (c *MessagingController) StartPatientConversationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := sharedhttp.BindAndValidate[dto.StartConversationBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	conv, err := c.uc.GetOrCreateConversationAsPatient(r.Context(), body.DoctorUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, conv)
}

func (c *MessagingController) GetOrCreateConversationHandler(w http.ResponseWriter, r *http.Request) {
	patientUUID, err := uuid.Parse(chi.URLParam(r, "patient_uuid"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	conv, err := c.uc.GetOrCreateConversation(r.Context(), patientUUID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, conv)
}

func (c *MessagingController) GetDoctorConversationsHandler(w http.ResponseWriter, r *http.Request) {
	convs, err := c.uc.GetDoctorConversations(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, convs)
}

func (c *MessagingController) GetPatientConversationsHandler(w http.ResponseWriter, r *http.Request) {
	convs, err := c.uc.GetPatientConversations(r.Context())
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}
	if convs == nil {
		convs = []dto.ConversationResponse{}
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, convs)
}

func (c *MessagingController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	convID, err := uuid.Parse(chi.URLParam(r, "conversation_id"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	body, err := sharedhttp.BindAndValidate[dto.SendDirectMessageBody](r, c.validator)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	msg, err := c.uc.SendMessage(r.Context(), convID, *body)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, msgStatusCode(err), err)
		return
	}

	sharedhttp.ResponseWithSuccess(w, r, http.StatusCreated, msg)
}

func (c *MessagingController) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	convID, err := uuid.Parse(chi.URLParam(r, "conversation_id"))
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusBadRequest, err)
		return
	}

	msgs, err := c.uc.GetMessages(r.Context(), convID)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, msgStatusCode(err), err)
		return
	}

	if msgs == nil {
		msgs = []dto.MessageResponse{}
	}
	sharedhttp.ResponseWithSuccess(w, r, http.StatusOK, msgs)
}

func msgStatusCode(err error) int {
	switch {
	case errors.Is(err, usecase.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, database.ErrConversationNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
