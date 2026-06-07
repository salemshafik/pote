package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
)

// NewRouter builds the chi router for the chat-service.
func NewRouter(h *ChatHandler, jwtSecret, frontendURL string) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack (matches other Pote services).
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1/chats", func(r chi.Router) {
		// All chat routes require a valid JWT.
		r.Use(authutils.AuthMiddleware(jwtSecret))

		r.Post("/", h.CreateChat)
		r.Get("/", h.ListChats)

		r.Route("/{chatID}", func(r chi.Router) {
			r.Get("/", h.GetChat)
			r.Put("/", h.UpdateChat)
			r.Delete("/", h.DeleteChat)

			r.Route("/members", func(r chi.Router) {
				r.Get("/", h.ListMembers)
				r.Post("/", h.AddMember)
				r.Delete("/me", h.Leave)
				r.Delete("/{userID}", h.RemoveMember)
				r.Put("/{userID}/role", h.UpdateMemberRole)
			})
		})
	})

	return r
}
