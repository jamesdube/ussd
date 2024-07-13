package ussd

import (
	"fmt"
	cfg "github.com/jamesdube/ussd/internal/config"
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/middleware"
	"github.com/jamesdube/ussd/pkg/router"
	"github.com/jamesdube/ussd/pkg/session"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log/slog"
)

type Framework struct {
	router             *router.Router
	registry           *gateway.Registry
	sessions           map[string]session.Session
	sessionRepository  session.Repository
	menuRegistry       *menu.Registry
	config             *config
	middlewareRegistry middleware.Registry
}

type config struct {
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

func Init(logger *slog.Logger) *Framework {

	utils.SetLogger(logger)
	configFile, err := ioutil.ReadFile("config.yaml")

	var c config
	if err != nil {
		utils.Logger.Warn("could not find config file")
	} else {
		err2 := yaml.Unmarshal(configFile, &c)
		if err2 != nil {
			utils.Logger.Warn("could not parse config file")
		}
	}

	sr := getRepository()

	f := &Framework{
		router:            router.NewRouter(),
		registry:          &gateway.Registry{},
		sessions:          map[string]session.Session{},
		menuRegistry:      menu.NewRegistry(),
		sessionRepository: sr,
		config:            &c,
	}

	f.setup()

	return f
}

func (f *Framework) GetGateway(s string) gateway.Gateway {
	return f.registry.Find(s)
}

func (f *Framework) GetSession(id string) (*session.Session, error) {

	utils.Logger.Debug("retrieving session [" + id + "] from repository")
	ss, err := f.sessionRepository.GetSession(id)
	if err != nil {
		return nil, err
	}
	return ss, nil
}

func (f *Framework) GetOrCreateSession(id string) (*session.Session, error) {

	utils.Logger.Debug("retrieving session [" + id + "] from repository")

	ss, err := f.GetSession(id)
	if err != nil {
		return nil, err
	}

	if ss == nil {
		return session.NewSession(id), nil
	}

	return ss, err

}

func (f *Framework) RemoveLastSessionEntry(id string) {
	ss, _ := f.GetSession(id)
	ss.RemoveLastSelection()
}

func (f *Framework) DeleteSession(id string) {
	utils.Logger.Debug("removing session [" + id + "] from repository")
	f.sessionRepository.Delete(id)
}

func (f *Framework) SaveSession(s *session.Session) {
	utils.Logger.Debug("saving session [" + s.Id + "] to repository")
	err := f.sessionRepository.Save(s)
	if err != nil {
		utils.Logger.Error(err.Error())
		return
	}

}

func (f *Framework) AddMenu(k string, m string) {
	mn := f.menuRegistry.Find(m)

	if mn != nil {
		f.router.AddRoute(k, mn)
		utils.Logger.Info("registered route", "routeKey", m, "routeMenu", m)
	}
}

func (f *Framework) setup() {
	e := gateway.NewEconetGateway()
	f.registry.Register(e)
}

func (f *Framework) configureMenus() {
	for k, v := range f.config.Menu.Navigation {
		f.AddMenu(k, v)
	}
}

func getRepository() session.Repository {

	p := cfg.Get("SESSION_PROVIDER")

	switch p {
	case "redis":
		logProvider("redis")
		return session.NewRedis()
	case "hazelcast":
		logProvider("hazelcast")
		return session.NewHazelCast("ussd")
	default:
		logProvider("memory")
		return session.NewInMemory()
	}
}

func logProvider(name string) {
	utils.Logger.Info("using session repository", "repository", name)
}

func loadConfig() {

	viper.SetConfigName("config")      // name of config file (without extension)
	viper.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/ussd/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.ussd") // call multiple times to add many search paths
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}
