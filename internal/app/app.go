package app

import (
	"net/http"
	"refer-mate/internal/config"
	routeHttp "refer-mate/internal/delivery/http"
	"refer-mate/internal/infrastrcture/gmail"
	"refer-mate/internal/infrastrcture/postgres"
	"refer-mate/internal/usecase"
)

type App struct {
	router http.Handler
}

func New(cfg *config.Config) (*App, error) {
	db, err := postgres.NewDB(cfg.DatabaseURL())
	if err != nil {
		return nil, err
	}

	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokenRepo(db)
	templateRepo := postgres.NewTemplateRepo(db)
	campaignRepo := postgres.NewCampaignRepo(db)
	recipientRepo := postgres.NewRecipientRepo(db)
	emailRepo := postgres.NewEmailRepo(db)

	gmailClient := gmail.NewClient(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Google.RedirectURL,
	)

	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, gmailClient, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	tmplUC := usecase.NewTemplateUseCase(templateRepo)
	campaignUC := usecase.NewCampaignUseCase(campaignRepo, recipientRepo, emailRepo, templateRepo, gmailClient, cfg.App.BaseURL)
	trackingUC := usecase.NewTrackingUseCase(emailRepo, campaignRepo)

	router := routeHttp.NewRouter(authUC, tmplUC, campaignUC, trackingUC)

	return &App{router: router}, nil
}

func (a *App) Run(addr string) error {
	return http.ListenAndServe(addr, a.router)
}
