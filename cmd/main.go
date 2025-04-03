package main

import (
	"log/slog"

	"lqkhoi-go-http-api/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := app.New()

	if err := app.Setup(); err != nil {
		slog.Error("Error when setting up server", "error", err)
	}
	app.Run()
}
