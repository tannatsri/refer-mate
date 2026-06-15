package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	"go-template/internal/usecase"
)

type EnvHandler struct {
	useCase *usecase.EnvUseCase
}

func NewEnvHandler(u *usecase.EnvUseCase) *EnvHandler {
	return &EnvHandler{useCase: u}
}

func registerEnvRouter(parent *mux.Router, routeUC *usecase.EnvUseCase) {

	envRouter := parent.PathPrefix("/env").Subrouter()

	handler := NewEnvHandler(routeUC)

	envRouter.HandleFunc("/list", handler.List).Methods("GET")
}

func (h *EnvHandler) List(w http.ResponseWriter, r *http.Request) {
	routes, err := h.useCase.ListRoutes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(routes)
}
