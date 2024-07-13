package ussd

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jamesdube/ussd/internal/utils"
	"log/slog"
)

func SetupLogging(app *fiber.App, logger *slog.Logger) {

	if logger != nil {
		utils.Logger = logger
		return
	}
	utils.Logger = slog.Default()
	//utils.InitializeLogger()
	//logger := utils.Logger

	//app.Use(fiberzap.New(fiberzap.Config{
	//	Logger: logger,
	//}))
}
