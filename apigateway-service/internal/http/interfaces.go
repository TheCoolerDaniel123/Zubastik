package http

import "net/http"

type AuthController interface {
	SendCodeHandler(w http.ResponseWriter, r *http.Request)
	VerifyCodeHandler(w http.ResponseWriter, r *http.Request)
	RefreshTokenHandler(w http.ResponseWriter, r *http.Request)
}

type UserController interface {
	GetProfileHandler(w http.ResponseWriter, r *http.Request)
	UpdateProfileHandler(w http.ResponseWriter, r *http.Request)
	GetTeethHandler(w http.ResponseWriter, r *http.Request)
	UpdateTeethHandler(w http.ResponseWriter, r *http.Request)
	GetBrushScheduleHandler(w http.ResponseWriter, r *http.Request)
	UpdateBrushScheduleHandler(w http.ResponseWriter, r *http.Request)
	GetDoctorsHandler(w http.ResponseWriter, r *http.Request)
}

type CalendarController interface {
	GetEventsHandler(w http.ResponseWriter, r *http.Request)
	CreateEventHandler(w http.ResponseWriter, r *http.Request)
	DeleteEventHandler(w http.ResponseWriter, r *http.Request)
}

type MessagingController interface {
	StartPatientConversationHandler(w http.ResponseWriter, r *http.Request)
	GetOrCreateConversationHandler(w http.ResponseWriter, r *http.Request)
	GetDoctorConversationsHandler(w http.ResponseWriter, r *http.Request)
	GetPatientConversationsHandler(w http.ResponseWriter, r *http.Request)
	SendMessageHandler(w http.ResponseWriter, r *http.Request)
	GetMessagesHandler(w http.ResponseWriter, r *http.Request)
}

type AiChatController interface {
	SendMessageHandler(w http.ResponseWriter, r *http.Request)
	GetHistoryHandler(w http.ResponseWriter, r *http.Request)
	GreetHandler(w http.ResponseWriter, r *http.Request)
}

type DoctorController interface {
	GetPatientProfileHandler(w http.ResponseWriter, r *http.Request)
	GetPatientTeethHandler(w http.ResponseWriter, r *http.Request)
	CreateToothNoteHandler(w http.ResponseWriter, r *http.Request)
	GetToothNotesHandler(w http.ResponseWriter, r *http.Request)
	UpdateToothNoteHandler(w http.ResponseWriter, r *http.Request)
	DeleteToothNoteHandler(w http.ResponseWriter, r *http.Request)
	CreateRecommendationHandler(w http.ResponseWriter, r *http.Request)
	GetRecommendationsHandler(w http.ResponseWriter, r *http.Request)
}

type ArticleController interface {
	GetApprovedArticlesHandler(w http.ResponseWriter, r *http.Request)
	CreateArticleHandler(w http.ResponseWriter, r *http.Request)
	GetAllArticlesHandler(w http.ResponseWriter, r *http.Request)
	ApproveArticleHandler(w http.ResponseWriter, r *http.Request)
	RejectArticleHandler(w http.ResponseWriter, r *http.Request)
}
