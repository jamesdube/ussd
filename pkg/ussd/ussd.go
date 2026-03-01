package ussd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	cfg "github.com/jamesdube/ussd/internal/config"
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/middleware"
	"github.com/jamesdube/ussd/pkg/router"
	"github.com/jamesdube/ussd/pkg/session"
	"gopkg.in/yaml.v3"
)

// --- Public API ---

// Ussd is the main entry point for the USSD framework.
type Ussd struct {
	framework *Framework
	config    *Config
}

// Config holds the application configuration.
type Config struct {
	AppName    string
	Port       int
	HideBanner bool
	Logger     *slog.Logger
}

// New creates a new Ussd instance.
func New(config ...Config) *Ussd {
	c := Config{}
	if len(config) > 0 {
		c = config[0]
	}
	return &Ussd{
		framework: newFramework(c.Logger),
		config:    &c,
	}
}

// AddMenu registers a menu implementation by name.
func (u *Ussd) AddMenu(name string, m menu.Menu) {
	u.framework.menuRegistry.Add(name, m)
}

// AddMiddleware registers a middleware handler.
func (u *Ussd) AddMiddleware(m middleware.Middleware) {
	u.framework.middlewareRegistry.Add(m)
}

// Start configures routes, metrics, and starts the HTTP server.
func (u *Ussd) Start() {
	app := fiber.New(fiber.Config{
		AppName:               u.config.AppName,
		DisableStartupMessage: u.config.HideBanner,
	})

	u.framework.configureMenus()
	app.Use(recover.New())

	// Routes
	app.Post("/econet", process(u.framework, "econet"))
	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	// Metrics
	svc := cfg.Get("APP_NAME")
	prometheus := fiberprometheus.New(svc)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	utils.Logger.Error(app.Listen(fmt.Sprintf(":%d", u.config.Port)).Error())
}

// --- Framework (internal) ---

// Framework holds all internal state for the USSD engine.
type Framework struct {
	router             *router.Router
	registry           *gateway.Registry
	sessionRepository  session.Repository
	menuRegistry       *menu.Registry
	config             *frameworkConfig
	middlewareRegistry middleware.Registry
}

type frameworkConfig struct {
	App struct {
		Port int
	}
	Cluster struct {
		Provider string
	}
	Menu struct {
		Navigation map[string]string
	}
}

// newFramework initializes the framework with config, session repository, and default gateway.
func newFramework(logger *slog.Logger) *Framework {
	utils.SetLogger(logger)

	var c frameworkConfig
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		utils.Logger.Warn("could not find config file")
	} else {
		if err := yaml.Unmarshal(configFile, &c); err != nil {
			utils.Logger.Warn("could not parse config file")
		}
	}

	reg := &gateway.Registry{}
	reg.Register(gateway.NewEconetGateway())

	return &Framework{
		router:            router.NewRouter(),
		registry:          reg,
		menuRegistry:      menu.NewRegistry(),
		sessionRepository: newSessionRepository(),
		config:            &c,
	}
}

func (f *Framework) GetGateway(s string) gateway.Gateway {
	return f.registry.Find(s)
}

func (f *Framework) GetSession(id string) (*session.Session, error) {
	utils.Logger.Debug("retrieving session", "sessionId", id)
	return f.sessionRepository.GetSession(id)
}

func (f *Framework) GetOrCreateSession(id string) (*session.Session, error) {
	ss, err := f.GetSession(id)
	if err != nil {
		return nil, err
	}
	if ss == nil {
		return session.NewSession(id), nil
	}
	return ss, nil
}

func (f *Framework) DeleteSession(id string) {
	utils.Logger.Debug("removing session", "sessionId", id)
	f.sessionRepository.Delete(id)
}

func (f *Framework) SaveSession(s *session.Session) {
	utils.Logger.Debug("saving session", "sessionId", s.Id)
	if err := f.sessionRepository.Save(s); err != nil {
		utils.Logger.Error(err.Error())
	}
}

func (f *Framework) AddMenu(k string, m string) {
	mn := f.menuRegistry.Find(m)
	if mn != nil {
		f.router.AddRoute(k, mn)
		utils.Logger.Debug("registered route", "routeKey", k, "routeMenu", m)
	}
}

func (f *Framework) configureMenus() {
	for k, v := range f.config.Menu.Navigation {
		f.AddMenu(k, v)
	}
}

func newSessionRepository() session.Repository {
	p := cfg.Get("SESSION_PROVIDER")
	switch p {
	case "redis":
		utils.Logger.Debug("using session repository", "repository", "redis")
		return session.NewRedis()
	case "hazelcast":
		utils.Logger.Debug("using session repository", "repository", "hazelcast")
		return session.NewHazelCast("ussd")
	default:
		utils.Logger.Debug("using session repository", "repository", "memory")
		return session.NewInMemory()
	}
}
