package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
)

// NewRouter creates and configures the chi router with all user-service routes.
func NewRouter(userHandler *UserHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Internal routes (service-to-service, no JWT auth)
	r.Route("/internal/v1", func(r chi.Router) {
		r.Post("/users/sync", userHandler.SyncProfile)
	})

	// Public routes (require JWT auth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authutils.AuthMiddleware(jwtSecret))

		// User profile routes
		r.Get("/users/me", userHandler.GetMe)
		r.Put("/users/me", userHandler.UpdateMe)
		r.Get("/users/search", userHandler.SearchUsers)
		r.Get("/users/{id}", userHandler.GetUser)

		// Contact routes
		r.Get("/contacts", userHandler.ListContacts)
		r.Post("/contacts", userHandler.AddContact)
		r.Delete("/contacts/{id}", userHandler.RemoveContact)

		// Invite routes
		r.Post("/invites", userHandler.SendInvite)
	})

	return r
}
