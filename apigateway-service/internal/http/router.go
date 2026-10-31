package http

import (
	"net/http"

	"apigateway-service/internal/http/middleware"
	"github.com/go-chi/chi"
	sharedjwt "shared/jwt"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(authController AuthController, userController UserController, calendarController CalendarController, doctorController DoctorController, messagingController MessagingController, chatController AiChatController, articleController ArticleController, jwtCfg *sharedjwt.JWTConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

	r.Route("/api/v1/auth", func(subr chi.Router) {
		subr.Post("/send_code", authController.SendCodeHandler)
		subr.Post("/verify_code", authController.VerifyCodeHandler)
		subr.Post("/refresh_token", authController.RefreshTokenHandler)
	})

	r.Route("/api/v1/profile", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))

		subr.Get("/", userController.GetProfileHandler)
		subr.Patch("/", userController.UpdateProfileHandler)

		subr.Get("/teeth", userController.GetTeethHandler)
		subr.Patch("/teeth", userController.UpdateTeethHandler)

		subr.Get("/brush-schedule", userController.GetBrushScheduleHandler)
		subr.Put("/brush-schedule", userController.UpdateBrushScheduleHandler)

		subr.Post("/conversations", messagingController.StartPatientConversationHandler)
		subr.Get("/conversations", messagingController.GetPatientConversationsHandler)
		subr.Get("/conversations/{conversation_id}/messages", messagingController.GetMessagesHandler)
		subr.Post("/conversations/{conversation_id}/messages", messagingController.SendMessageHandler)
	})

	r.Route("/api/v1/calendar", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Get("/events", calendarController.GetEventsHandler)
		subr.Post("/events", calendarController.CreateEventHandler)
		subr.Delete("/events/{event_id}", calendarController.DeleteEventHandler)
	})

	r.Route("/api/v1/ai-chat", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Post("/message", chatController.SendMessageHandler)
		subr.Get("/history", chatController.GetHistoryHandler)
		subr.Post("/greet", chatController.GreetHandler)
	})

	r.Route("/api/v1/doctor", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Use(middleware.RequireRole(1)) // RoleDoctor
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
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Get("/", userController.GetDoctorsHandler)
	})

	r.Route("/api/v1/lectures", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Get("/", articleController.GetApprovedArticlesHandler)
	})

	r.Route("/api/v1/doctor/lectures", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Use(middleware.RequireRole(1))
		subr.Post("/", articleController.CreateArticleHandler)
	})

	r.Route("/api/v1/admin/lectures", func(subr chi.Router) {
		subr.Use(middleware.JWTMiddleware(jwtCfg))
		subr.Use(middleware.RequireRole(2))
		subr.Get("/", articleController.GetAllArticlesHandler)
		subr.Patch("/{article_id}/approve", articleController.ApproveArticleHandler)
		subr.Patch("/{article_id}/reject", articleController.RejectArticleHandler)
	})

	return r
}
