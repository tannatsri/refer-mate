package http

import (
	"net/http"
	"net/url"
	"refer-mate/internal/usecase"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	authUC      *usecase.AuthUseCase
	frontendURL string
}

func NewAuthHandler(authUC *usecase.AuthUseCase, frontendURL string) *AuthHandler {
	return &AuthHandler{authUC: authUC, frontendURL: frontendURL}
}

func registerAuthRoutes(parent *mux.Router, authUC *usecase.AuthUseCase, frontendURL string) {
	h := NewAuthHandler(authUC, frontendURL)
	auth := parent.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/google", h.GoogleLogin).Methods("GET")
	auth.HandleFunc("/google/callback", h.GoogleCallback).Methods("GET")
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "random-state"
	url := h.authUC.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	jwtToken, _, err := h.authUC.HandleCallback(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	redirectURL := h.frontendURL + "/auth/callback?token=" + url.QueryEscape(jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
