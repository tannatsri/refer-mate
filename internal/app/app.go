package app

import (
	"go-template/internal/config"
	routeHttp "go-template/internal/delivery/http"
	"go-template/internal/infrastrcture/postgres"
	"go-template/internal/usecase"
	"net/http"
)

type App struct {
	router http.Handler
}

func New(cfg *config.Config) (*App, error) {

	dbConn, err := postgres.NewDB(cfg.DatabaseURL())
	if err != nil {
		return nil, err
	}

	envRepo := postgres.NewEnvRepo(dbConn)
	envUC := usecase.NewEnvUseCase(envRepo)
	router := routeHttp.NewRouter(envUC)

	return &App{router: router}, nil
}

func (a *App) Run(addr string) error {
	return http.ListenAndServe(addr, a.router)
}
