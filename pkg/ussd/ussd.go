package ussd

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/middleware"
	"log/slog"
)

type Ussd struct {
	framework *Framework
	config    *Config
}

type Config struct {
	AppName    string
	Port       int
	HideBanner bool
	Logger     *slog.Logger
}

func New(config ...Config) *Ussd {

	cfg := Config{}

	if len(config) > 0 {
		cfg = config[0]
	}

	return &Ussd{
		framework: Init(cfg.Logger),
		config:    &cfg,
	}
}

func (u *Ussd) AddMenu(name string, m menu.Menu) {
	u.framework.menuRegistry.Add(name, m)
}

func (u *Ussd) AddMiddleware(m middleware.Middleware) {
	u.framework.middlewareRegistry.Add(m)
}

func (u *Ussd) Start() {

	app := fiber.New(fiber.Config{
		AppName:               u.config.AppName,
		DisableStartupMessage: u.config.HideBanner,
	})

	u.framework.configureMenus()

	app.Use(recover.New())

	SetupRoutes(u.framework, app)
	SetupMetrics(app)
	//utils.SetLogger(u.logger)

	utils.Logger.Error(app.Listen(fmt.Sprintf(":%d", u.config.Port)).Error())

}
