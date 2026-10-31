package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	sharedhttp "shared/http"
	sharedjwt "shared/jwt"
)

var errInvalidToken = errors.New("invalid or expired token")

func jwtMiddleware(cfg *sharedjwt.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			claims, err := sharedjwt.ParseAndValidateAccessToken(cfg, tokenStr)
			if err != nil {
				sharedhttp.ResponseWithError(w, r, http.StatusUnauthorized, errInvalidToken)
				return
			}
			ctx := context.WithValue(r.Context(), sharedjwt.CtxKeyClaims{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewRouter(userController UserController, toothController ToothController, brushController BrushController, calendarController CalendarController, doctorController DoctorController, messagingController MessagingController, chatController AiChatController, articleController ArticleController, doctorsListController DoctorsListController, jwtCfg *sharedjwt.JWTConfig) http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1/profile", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/", userController.GetProfileHandler)
		subr.Patch("/", userController.UpdateProfileHandler)

		subr.Get("/teeth", toothController.GetTeethHandler)
		subr.Patch("/teeth", toothController.UpdateTeethHandler)

		subr.Get("/brush-schedule", brushController.GetBrushScheduleHandler)
		subr.Put("/brush-schedule", brushController.UpdateBrushScheduleHandler)

		subr.Post("/conversations", messagingController.StartPatientConversationHandler)
		subr.Get("/conversations", messagingController.GetPatientConversationsHandler)
		subr.Get("/conversations/{conversation_id}/messages", messagingController.GetMessagesHandler)
		subr.Post("/conversations/{conversation_id}/messages", messagingController.SendMessageHandler)
	})

	r.Route("/api/v1/calendar", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/events", calendarController.GetCalendarEventsHandler)
		subr.Post("/events", calendarController.CreateCalendarEventHandler)
		subr.Delete("/events/{event_id}", calendarController.DeleteCalendarEventHandler)
	})

	r.Route("/api/v1/ai-chat", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Post("/message", chatController.SendMessageHandler)
		subr.Get("/history", chatController.GetHistoryHandler)
		subr.Post("/greet", chatController.GreetHandler)
	})

	r.Route("/api/v1/doctor", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/patients/{patient_uuid}", doctorController.GetPatientProfileHandler)
		subr.Get("/patients/{patient_uuid}/teeth", doctorController.GetPatientTeethHandler)
		subr.Post("/patients/{patient_uuid}/teeth/notes", doctorController.CreateToothNoteHandler)
		subr.Get("/patients/{patient_uuid}/teeth/notes", doctorController.GetToothNotesHandler)
		subr.Put("/patients/{patient_uuid}/teeth/notes/{note_id}", doctorController.UpdateToothNoteHandler)
		subr.Delete("/patients/{patient_uuid}/teeth/notes/{note_id}", doctorController.DeleteToothNoteHandler)
		subr.Post("/patients/{patient_uuid}/recommendations", doctorController.CreateRecommendationHandler)
		subr.Get("/patients/{patient_uuid}/recommendations", doctorController.GetRecommendationsHandler)

		subr.Post("/patients/{patient_uuid}/conversations", messagingController.GetOrCreateConversationHandler)
		subr.Get("/conversations", messagingController.GetDoctorConversationsHandler)
		subr.Get("/conversations/{conversation_id}/messages", messagingController.GetMessagesHandler)
		subr.Post("/conversations/{conversation_id}/messages", messagingController.SendMessageHandler)
	})

	r.Route("/api/v1/doctors", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/", doctorsListController.GetDoctorsHandler)
	})

	r.Route("/api/v1/lectures", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/", articleController.GetApprovedArticlesHandler)
	})

	r.Route("/api/v1/doctor/lectures", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Post("/", articleController.CreateArticleHandler)
	})

	r.Route("/api/v1/admin/lectures", func(subr chi.Router) {
		subr.Use(jwtMiddleware(jwtCfg))
		subr.Get("/", articleController.GetAllArticlesHandler)
		subr.Patch("/{article_id}/approve", articleController.ApproveArticleHandler)
		subr.Patch("/{article_id}/reject", articleController.RejectArticleHandler)
	})

	return r
}
