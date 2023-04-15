package ussd

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	cfg "github.com/jamesdube/ussd/internal/config"
)

func SetupMetrics(app *fiber.App) {

	svc := cfg.Get("APP_NAME")

	prometheus := fiberprometheus.New(svc)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
}
