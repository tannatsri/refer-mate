package main

import (
	"log"

	"refer-mate/internal/app"
	"refer-mate/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server running on port " + cfg.App.Port)
	log.Fatal(application.Run(":" + cfg.App.Port))
}
