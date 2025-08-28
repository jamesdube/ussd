package gateway

import "github.com/gofiber/fiber/v2"

type Registry struct {
	gateways []Gateway
}

type Gateway interface {
	ToRequest(b *fiber.Ctx) (Request, error)
	Request() Request
	ToResponse(response Response) interface{}
	Name() string
}

type Request struct {
	SessionId string
	Message   string
	Msisdn    string
	Stage	  string
	DestinationNumber string // might change name later
}

type Response struct {
	Message       string
	Session       string
	Msisdn        string
	SessionActive bool
}

func (r *Registry) Register(g Gateway) {
	r.gateways = append(r.gateways, g)
}

func (r *Registry) Find(n string) Gateway {

	for _, gateway := range r.gateways {
		if gateway.Name() == n {
			return gateway
		}
	}
	return nil
}

func NewRegistry() *Registry {
	return &Registry{gateways: nil}
}
