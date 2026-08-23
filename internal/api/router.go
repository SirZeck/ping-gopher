package api

import (
	"net/http"

	"github.com/SirZeck/ping-gopher/web"
)

// SetupRouter initializes the HTTP router and registers all public and authenticated REST endpoints.
func (h *APIHandler) SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// 1. Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		JSONResponse(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "ping-gopher-api",
		})
	})

	// 2. Authentication endpoints (Public)
	mux.HandleFunc("POST /v1/auth/signup", h.SignupHandler)
	mux.HandleFunc("POST /v1/auth/login", h.LoginHandler)

	// 3. Public Status Page Data Endpoint
	mux.HandleFunc("GET /v1/status/public", h.PublicStatusHandler)

	// 4. Monitor management endpoints (JWT Authenticated)
	jwtSecret := h.Config.JWTSecret
	mux.HandleFunc("POST /v1/monitors", AuthMiddleware(jwtSecret, h.CreateMonitorHandler))
	mux.HandleFunc("GET /v1/monitors", AuthMiddleware(jwtSecret, h.ListMonitorsHandler))
	mux.HandleFunc("GET /v1/monitors/{id}", AuthMiddleware(jwtSecret, h.GetMonitorHandler))
	mux.HandleFunc("PUT /v1/monitors/{id}", AuthMiddleware(jwtSecret, h.UpdateMonitorHandler))
	mux.HandleFunc("DELETE /v1/monitors/{id}", AuthMiddleware(jwtSecret, h.DeleteMonitorHandler))

	// 5. Telemetry logs & incident history (JWT Authenticated)
	mux.HandleFunc("GET /v1/monitors/{id}/logs", AuthMiddleware(jwtSecret, h.GetMonitorLogsHandler))
	mux.HandleFunc("GET /v1/monitors/{id}/incidents", AuthMiddleware(jwtSecret, h.GetMonitorIncidentsHandler))

	// 6. Embedded Web Dashboard Static Handler
	mux.Handle("GET /", web.StaticHandler())

	// Wrap in CORS middleware
	return CORSMiddleware(mux)
}
