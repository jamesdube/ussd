package ussd

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(framework *Framework, app *fiber.App) {

	app.Post("/econet", handle(framework, "econet"))
	app.Get("/health", health)

}

func handle(f *Framework, gn string) func(ctx *fiber.Ctx) error {
	return process(f, gn)
}

func health(ctx *fiber.Ctx) error {
	return ctx.SendString("ok")
}
