package router

import (
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/menu"
	"go.uber.org/zap"
)

type Router struct {
	menus  []menu.Menu
	routes map[string]menu.Menu
}

func NewRouter() *Router {
	return &Router{
		menus:  []menu.Menu{},
		routes: map[string]menu.Menu{},
	}
}

func (r *Router) AddRoute(route string, menu menu.Menu) {
	r.routes[route] = menu
}

func (r *Router) RouteTo(s []string) menu.Menu {

	for k, m := range r.routes {
		route := utils.StringToSlice(k)

		found := check(route, s)
		if found {
			utils.Logger.Info("routing to ", zap.String("route", k))
			return m
		}
	}
	return nil
}

func check(k []string, v []string) bool {

	if len(k) != len(v) {
		return false
	}

	for i := 0; i < len(k); i++ {
		match := comparator(k[i], v[i])

		if !match {
			return false
		}
	}

	return true
}

func comparator(v1 string, v2 string) bool {

	if v1 == v2 {
		return true
	}

	if v1 == "*" {
		return true
	}

	return false
}
