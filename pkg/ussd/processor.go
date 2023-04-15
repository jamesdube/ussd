package ussd

import (
	"encoding/xml"
	"github.com/gofiber/fiber/v2"
	u "github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/session"
	"go.uber.org/zap"
	"strings"
)

func process(framework *Framework, name string) func(ctx *fiber.Ctx) error {

	gw := framework.GetGateway(name)

	return func(ctx *fiber.Ctx) error {

		gr, err := gw.ToRequest(ctx)

		if err != nil {
			u.Logger.Error("Can't unmarshal the byte array")
			return ctx.SendString("failed to unmarshal")
		}

		msg := gr.Message

		ss, e := framework.GetOrCreateSession(gr.SessionId)

		if e != nil {
			u.Logger.Error("failed to initiate session")
			return onError(framework, ctx, gw, ss, gr.Msisdn)
		}

		err = runMiddleware(framework, ss, gr)
		if err != nil {
			return onError(framework, ctx, gw, ss, gr.Msisdn)
		}

		ss.AddSelection(msg)
		framework.SaveSession(ss)
		mn := framework.router.RouteTo(ss.GetSelections())

		if mn == nil {
			u.Logger.Error("menu not found for route", zap.Any("route", ss.GetSelections()))
			return onErrorWith(u.MenuInvalidSelection, framework, ctx, gw, ss, gr.Msisdn)
		}

		c := menu.NewContext(gr.Msisdn, ss.Attributes)
		response := mn.Render(c, msg)

		postNavigation(framework, c, ss)

		r := buildResponse(gw, response, ss, gr.Msisdn, c.Active)
		return sendResponse(r, ctx)
	}

}

func onError(framework *Framework, ctx *fiber.Ctx, gateway gateway.Gateway, ss *session.Session, msisdn string) error {

	framework.DeleteSession(ss.Id)
	r := buildResponse(gateway, u.MenuInvalidSelection, ss, msisdn, false)
	return sendResponse(r, ctx)

}

func onErrorWith(msg string, framework *Framework, ctx *fiber.Ctx, gateway gateway.Gateway, ss *session.Session, msisdn string) error {

	u.Logger.Error(msg)
	framework.DeleteSession(ss.Id)
	r := buildResponse(gateway, u.MenuInvalidSelection, ss, msisdn, false)
	return sendResponse(r, ctx)

}

func postNavigation(f *Framework, c *menu.Context, ss *session.Session) {

	switch c.NavigationType {

	case menu.Stop:
		{
			f.DeleteSession(ss.Id)
			c.Active = false
		}
	case menu.Continue:
		{
			f.SaveSession(ss)
		}
	case menu.Replay:
		{
			f.RemoveLastSessionEntry(ss.Id)
			f.RemoveLastSessionEntry(ss.Id)
			f.SaveSession(ss)
		}

	}
}

func runMiddleware(f *Framework, ss *session.Session, gr gateway.Request) error {
	for _, m := range f.middlewareRegistry.Get() {
		errM := m.Handle(ss, &gr)
		if errM != nil {
			return errM
		}
	}
	return nil
}

func buildResponse(g gateway.Gateway, message string, session *session.Session, msisdn string, active bool) interface{} {

	return g.ToResponse(gateway.Response{
		Message:       message,
		Session:       session.GetID(),
		Msisdn:        msisdn,
		SessionActive: active,
	})

}

func sendResponse(grs interface{}, ctx *fiber.Ctx) error {

	result, _ := xml.Marshal(&grs)
	xmls := strings.ReplaceAll(string(result), "&#xA;", "\n")

	ctx.Type("xml")
	return ctx.Send([]byte(u.Header + xmls))
}
