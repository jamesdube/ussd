package ussd

import (
	"github.com/gofiber/fiber/v2"
	u "github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/session"
)

type ProcessHandler interface {
	Render(message string, ctx *menu.Context, sess *session.Session, f *Framework) menu.Response
}

type Request struct {
	Msisdn  string
	Message string
}

type Response struct {
	Message       string
	Session       string
	Msisdn        string
	SessionActive bool
}

type ProcessHandlerManager struct {
}

func (phm *ProcessHandlerManager) handle(ctx *menu.Context) ProcessHandler {
	switch ctx.NavigationType {

	case menu.Continue:
		return &Normal{}

	case menu.Paginated:
		return &Paginated{}
	}

	return nil
}

func p(f *Framework, name string) func(ctx *fiber.Ctx) error {

	gw := gateway.NewEconetGateway()

	return func(ctx *fiber.Ctx) error {

		gr, err := gw.ToRequest(ctx)

		if err != nil {
			u.Logger.Error("Can't unmarshal the byte array")
			return ctx.SendString("failed to unmarshal")
		}

		r := Request{Message: gr.Message}

		res := h(r, &ProcessHandlerManager{}, f)

		gwr := gw.ToResponse(gateway.Response{
			Message:       res.Message,
			Session:       res.Session,
			Msisdn:        res.Msisdn,
			SessionActive: true,
		})

		return sendResponse(gwr, ctx)
	}

}

func h(req Request, phm *ProcessHandlerManager, f *Framework) Response {

	sess, _ := f.GetOrCreateSession(req.Msisdn)
	c := menu.NewContext(req.Msisdn, sess)

	ph := phm.handle(c)

	mr := ph.Render(req.Message, c, sess, f)

	r := build(req, mr, sess)

	return r

}

func build(req Request, mr menu.Response, sess *session.Session) Response {

	m := mr.Prompt

	if len(mr.Options) > 0 {
		opt := buildOptions(mr.Options)
		m = m + opt
	}

	return Response{
		Message:       m,
		Session:       sess.GetID(),
		Msisdn:        req.Msisdn,
		SessionActive: true,
	}
}
