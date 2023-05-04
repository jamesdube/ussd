package ussd

import (
	"encoding/xml"
	"fmt"
	"github.com/gofiber/fiber/v2"
	u "github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/session"
	"go.uber.org/zap"
	"strconv"
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
			return onErrorWith(err.Error(), framework, ctx, gw, ss, gr.Msisdn)
		}

		c := menu.NewContext(gr.Msisdn, ss)

		if c.Paginated {

			fmt.Println("pagination wanted")
			return handlePagination(framework, c, ctx, gr.Message, "from context session", gr.Msisdn, gw, ss)

		}

		prev := framework.router.RouteTo(ss.GetSelections())

		if prev != nil {
			prev.Process(c, msg)
		}

		if c.NavigationType == menu.Replay {

			fmt.Println("replay wanted")
			pr := prev.OnRequest(c, msg)

			postNavigation(framework, c, ss)

			r := buildResponse(gw, pr.Prompt, pr.Options, ss, gr.Msisdn, c.Active)
			return sendResponse(r, ctx)

		}

		ss.AddSelection(msg)
		framework.SaveSession(ss)
		mn := framework.router.RouteTo(ss.GetSelections())

		if mn == nil {
			u.Logger.Error("menu not found for route", zap.Any("route", ss.GetSelections()))
			return onErrorWith(u.MenuInvalidSelection, framework, ctx, gw, ss, gr.Msisdn)
		}

		rMsg := mn.OnRequest(c, msg)

		if c.Paginated {

			createPagination(c, rMsg, ss)

			postNavigation(framework, c, ss)

			return handlePagination(framework, c, ctx, gr.Message, rMsg.Prompt, gr.Msisdn, gw, ss)

		}

		postNavigation(framework, c, ss)

		r := buildResponse(gw, rMsg.Prompt, rMsg.Options, ss, gr.Msisdn, c.Active)
		return sendResponse(r, ctx)
	}

}

func onError(framework *Framework, ctx *fiber.Ctx, gateway gateway.Gateway, ss *session.Session, msisdn string) error {

	framework.DeleteSession(ss.Id)
	r := buildResponse(gateway, u.MenuInvalidSelection, nil, ss, msisdn, false)
	return sendResponse(r, ctx)

}

func onErrorWith(msg string, framework *Framework, ctx *fiber.Ctx, gateway gateway.Gateway, ss *session.Session, msisdn string) error {

	u.Logger.Error(msg)
	framework.DeleteSession(ss.Id)
	r := buildResponse(gateway, u.MenuInvalidSelection, nil, ss, msisdn, false)
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

func buildResponse(g gateway.Gateway, message string, options []string, session *session.Session, msisdn string, active bool) interface{} {

	m := message

	if len(options) > 0 {
		opt := buildOptions(options)
		m = message + opt
	}

	if session.Paginated {
		m = m + "\n\n 0 More..."
	}

	return g.ToResponse(gateway.Response{
		Message:       m,
		Session:       session.GetID(),
		Msisdn:        msisdn,
		SessionActive: active,
	})

}

func buildOptions(options []string) string {

	var sb strings.Builder
	i := 1
	for _, o := range options {
		sb.WriteString("\n" + strconv.Itoa(i) + ". " + o)
		i++
	}
	return sb.String()
}

func sendResponse(grs interface{}, ctx *fiber.Ctx) error {

	result, _ := xml.Marshal(&grs)
	xmls := strings.ReplaceAll(string(result), "&#xA;", "\n")

	ctx.Type("xml")
	return ctx.Send([]byte(u.Header + xmls))
}

func handlePagination(framework *Framework, c *menu.Context, ctx *fiber.Ctx, message string, prompt string, msisdn string, gateway gateway.Gateway, session *session.Session) error {

	first := session.CurrentPage == 0
	cont := first || message == "0"
	last := (len(c.Pages)) == 1 || (len(c.Pages)) == (c.CurrentPage)

	fmt.Println("pagination option processing menu")

	if !first && !cont || last {

		io, e := strconv.Atoi(message)
		validOption := isValidOption(c, io)
		if e != nil || !validOption {
			postNavigation(framework, c, session)
			u.Logger.Error("invalid pagination option", zap.Any("route", session.GetSelections()))
			return onErrorWith(u.MenuInvalidSelection, framework, ctx, gateway, session, msisdn)
		}

		var count int

		if (len(c.Pages)) > 1 && c.CurrentPage > 1 {
			pct := len(c.Pages[c.CurrentPage-2])
			count = pct
		}

		c.SelectedPaginationOption = io + count
		prev := framework.router.RouteTo(session.GetSelections())
		prev.Process(c, message)

		session.AddSelection(message)
		mn := framework.router.RouteTo(session.GetSelections())

		if mn == nil {

			u.Logger.Error("menu not found for route", zap.Any("route", session.GetSelections()))
			return onErrorWith(u.MenuInvalidSelection, framework, ctx, gateway, session, msisdn)
		}

		res := mn.OnRequest(c, message)

		postNavigation(framework, c, session)

		r := buildResponse(gateway, res.Prompt, res.Options, session, msisdn, c.Active)
		return sendResponse(r, ctx)

	}

	if cont {
		session.CurrentPage++
		framework.SaveSession(session)
	}

	/*	if (len(c.Pages) - 1) == (c.CurrentPage) {
			fmt.Println("Last Page")
			session.Paginated = false
			//framework.RemoveLastSessionEntry(session.Id)

			mn := framework.router.RouteTo(session.GetSelections())

			fmt.Println("pagination option processing menu")
			mn.Process(c, message)

			postNavigation(framework, c, session)

		} else {
			fmt.Println("Not Last Page")
			session.CurrentPage++
			framework.SaveSession(session)
		}*/

	r := buildResponse(gateway, prompt, c.Pages[c.CurrentPage], session, msisdn, c.Active)

	return sendResponse(r, ctx)

}

func isValidOption(ctx *menu.Context, io int) bool {

	if ctx.CurrentPage == 0 {
		return true
	}

	b := (io - 1) < len(ctx.Pages[ctx.CurrentPage-1])
	return b
}

func createPagination(c *menu.Context, menuResponse menu.Response, session *session.Session) {

	if menuResponse.PerPage == 0 {
		menuResponse.PerPage = len(menuResponse.Options)
	}

	var pages [][]string
	//var options []string
	for i := 0; i < len(menuResponse.Options); i = i + menuResponse.PerPage {

		r := (i + menuResponse.PerPage) < len(menuResponse.Options)
		max := i + menuResponse.PerPage
		if !r {
			max = len(menuResponse.Options)
		}

		pages = append(pages, menuResponse.Options[i:max])

	}

	c.Pages = pages
	c.CurrentPage = 0
	session.Paginated = true
	session.Pages = c.Pages
	session.CurrentPage = c.CurrentPage

}
