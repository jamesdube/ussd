package ussd

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jamesdube/ussd/internal/utils"
)

func SetupLogging(app *fiber.App) {

	utils.InitializeLogger()
	//logger := utils.Logger

	//app.Use(fiberzap.New(fiberzap.Config{
	//	Logger: logger,
	//}))
}
