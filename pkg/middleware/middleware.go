package middleware

import (
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/session"
)

type Middleware interface {
	Handle(s *session.Session, gr *gateway.Request) error
}

type Registry struct {
	middleware []Middleware
}

func (r *Registry) Add(m Middleware) {
	r.middleware = append(r.middleware, m)
}

func (r *Registry) Get() []Middleware {
	return r.middleware
}
