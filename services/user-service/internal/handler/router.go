package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
)

// NewRouter creates and configures the chi router with all user-service routes.
// Protected routes are guarded by the shared authutils JWT middleware.
func NewRouter(userHandler *UserHandler, jwtSecret, frontendURL string) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1/users", func(r chi.Router) {
		// Internal provisioning (service-to-service). Created without the
		// user JWT because the user does not exist in this service yet.
		r.Post("/", userHandler.CreateProfile)

		// Protected routes — require a valid access token.
		r.Group(func(r chi.Router) {
			r.Use(authutils.AuthMiddleware(jwtSecret))

			// Self profile
			r.Get("/me", userHandler.GetMe)
			r.Put("/me", userHandler.UpdateMe)
			r.Put("/me/status", userHandler.UpdateStatus)

			// Contacts
			r.Get("/me/contacts", userHandler.ListContacts)
			r.Post("/me/contacts", userHandler.AddContact)
			r.Delete("/me/contacts/{contactID}", userHandler.RemoveContact)

			// Invites
			r.Get("/me/invites", userHandler.ListInvites)
			r.Post("/me/invites", userHandler.CreateInvite)

			// Public profile lookup by ID (kept last so it doesn't shadow /me).
			r.Get("/{id}", userHandler.GetProfile)
		})
	})

	return r
}
