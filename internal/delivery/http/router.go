package http

import (
	"go-template/internal/delivery/http/middlewear"
	"net/http"

	"github.com/gorilla/mux"
	"go-template/internal/usecase"
)

func NewRouter(envUC *usecase.EnvUseCase) http.Handler {
	r := mux.NewRouter()

	// Global middleware
	r.Use(middlewear.Recovery)

	api := r.PathPrefix("/api/v1").Subrouter()

	registerEnvRouter(api, envUC)

	return r
}
