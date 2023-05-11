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

	if session.PaginatedHasMore {
		m = m + "\n0. More"
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
	last := (len(c.Pages)) == (c.CurrentPage) || (len(c.Pages)) == 1

	if len(c.Pages) == (c.CurrentPage + 1) {
		fmt.Println("last page")
		session.PaginatedHasMore = false
	}

	if last && message == "0" {
		return onErrorWith(u.MenuNoMoreOptions, framework, ctx, gateway, session, msisdn)
	}

	fmt.Println("pagination option processing menu")

	fmt.Println("current page original:", c.CurrentPage)
	fmt.Println("current page calculated:", c.CurrentPage-1)
	//fmt.Println("current pages:", c.Pages[c.CurrentPage])

	if !first && !cont || last {

		//c.Paginated = false
		//session.Paginated = false

		io, e := strconv.Atoi(message)
		validOption := isValidOption(c, io)
		if e != nil || !validOption {
			postNavigation(framework, c, session)
			u.Logger.Error("invalid pagination option", zap.Any("route", session.GetSelections()))
			return onErrorWith(u.MenuInvalidSelection, framework, ctx, gateway, session, msisdn)
		}

		var optionsCount int
		if (len(c.Pages)) > 1 && c.CurrentPage > 1 {
			pagesViewed := c.CurrentPage - 1
			for i := 0; i < pagesViewed; i++ {
				optionsCount = optionsCount + len(c.Pages[i])
			}
		}

		fmt.Println("options count:", optionsCount)

		c.SelectedPaginationOption = io + optionsCount
		c.SelectedPageOption = io
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

		c.Paginated = false
		session.Paginated = false

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
	session.PaginatedHasMore = len(c.Pages) > 1
	session.Pages = c.Pages
	session.CurrentPage = c.CurrentPage

}
