package ussd

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/middleware"
	"log"
)

type Ussd struct {
	framework *Framework
	port      int
}

func New() *Ussd {
	return &Ussd{
		framework: Init(),
		port:      7600,
	}
}

func (u *Ussd) AddMenu(name string, m menu.Menu) {
	u.framework.menuRegistry.Add(name, m)
}

func (u *Ussd) AddMiddleware(m middleware.Middleware) {
	u.framework.middlewareRegistry.Add(m)
}

func (u *Ussd) SetPort(port int) {
	u.port = port
}

func (u *Ussd) Start() {

	u.framework.configureMenus()
	app := fiber.New()

	app.Use(recover.New())

	//SetupLogging(app)
	SetupRoutes(u.framework, app)
	SetupMetrics(app)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", u.port)))

}
