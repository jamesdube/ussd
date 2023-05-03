package main

import (
	"github.com/jamesdube/ussd/internal/config"
	"github.com/jamesdube/ussd/m"
	"github.com/jamesdube/ussd/pkg/ussd"
	_ "github.com/jamesdube/ussd/pkg/ussd"
	"strconv"
)

func main() {

	u := ussd.New()
	l := &m.Login{}
	h := &m.Home{}
	u.AddMenu("login", l)
	u.AddMenu("home", h)
	u.AddMenu("foo", &m.Foo{})
	u.AddMenu("page", &m.Paginated{})
	u.AddMenu("opt", &m.Option{})

	p := config.Get("APP_PORT")
	pi := 7600

	if p != "" {
		pi, _ = strconv.Atoi(p)
	}

	//u.AddMiddleware(&m.Auth{})

	u.SetPort(pi)
	u.Start()

}
