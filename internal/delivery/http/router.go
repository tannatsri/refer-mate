package http

import (
	"encoding/json"
	"net/http"
	"refer-mate/internal/delivery/http/middlewear"
	"refer-mate/internal/usecase"

	"github.com/gorilla/mux"
)

func NewRouter(
	authUC *usecase.AuthUseCase,
	tmplUC *usecase.TemplateUseCase,
	campaignUC *usecase.CampaignUseCase,
	trackingUC *usecase.TrackingUseCase,
	frontendURL string,
) http.Handler {
	r := mux.NewRouter()
	r.Use(middlewear.Recovery)
	r.Use(middlewear.CORS(frontendURL))

	// gorilla/mux only invokes Use()-registered middleware for routes that
	// match; without an explicit OPTIONS route, preflight requests fall
	// through to mux's 404 handler and never get CORS headers.
	r.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Tracking routes (no auth required)
	registerTrackingRoutes(r, trackingUC)

	api := r.PathPrefix("/api/v1").Subrouter()

	// Public auth routes
	registerAuthRoutes(api, authUC, frontendURL)

	// Protected routes
	protected := api.NewRoute().Subrouter()
	protected.Use(middlewear.Auth(authUC))

	protected.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		userID := middlewear.GetUserID(r)
		user, err := authUC.GetMe(r.Context(), userID)
		if err != nil || user == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}).Methods("GET")

	registerTemplateRoutes(protected, tmplUC)
	registerCampaignRoutes(protected, campaignUC, authUC)

	return r
}
