package ussd

//import (
//	"fmt"
//	u "github.com/jamesdube/ussd/internal/utils"
//	"github.com/jamesdube/ussd/pkg/gateway"
//	"github.com/jamesdube/ussd/pkg/menu"
//	"github.com/jamesdube/ussd/pkg/session"
//	"go.uber.org/zap"
//	"strconv"
//	"strings"
//)
//
//func h(framework *Framework, gateway string, request gateway.Request) interface{} {
//
//	gw := framework.GetGateway(gateway)
//	ss, e := getOrCreateSession(request)
//	ctx := newContext()
//	prev := framework.router.RouteTo(ss.GetSelections())
//
//	if e != nil {
//		u.Logger.Error("failed to initiate session")
//		return errorResponse(framework, ctx, gw, ss, request.Msisdn)
//	}
//
//	switch ctx.NavigationType {
//
//	case menu.Continue:
//
//		ss.AddSelection(request.Message)
//		framework.SaveSession(ss)
//		mn := framework.router.RouteTo(ss.GetSelections())
//
//		if mn == nil {
//			u.Logger.Error("menu not found for route", zap.Any("route", ss.GetSelections()))
//			return errorResponse(framework, ctx, gw, ss, "abc")
//		}
//
//		mnResp := mn.OnRequest(ctx, request.Message)
//
//		return response(gw, mnResp.Prompt, mnResp.Options, ss, request.Msisdn, ctx.Active)
//
//	case menu.Replay:
//		{
//
//			fmt.Println("replay wanted")
//
//			pr := prev.OnRequest(ctx, request.Message)
//
//			postNavigation(framework, ctx, ss)
//
//			return response(gw, pr.Prompt, pr.Options, ss, request.Msisdn, ctx.Active)
//
//		}
//
//	case menu.Stop:
//		{
//			postNavigation(framework, ctx, ss)
//		}
//
//	case menu.Paginated:
//
//	case menu.LongCode:
//		{
//
//		}
//
//	}
//
//	menuOptionResponse()
//
//	if ctx.IsReplay() {
//		menuReplay()
//	}
//
//}
//
//func response(g gateway.Gateway, message string, options []string, session *session.Session, msisdn string, active bool) interface{} {
//
//	m := message
//
//	if len(options) > 0 {
//		opt := buildOptions(options)
//		m = message + opt
//	}
//
//	return g.ToResponse(gateway.Response{
//		Message:       m,
//		Session:       session.GetID(),
//		Msisdn:        msisdn,
//		SessionActive: active,
//	})
//
//}
//
//func options(options []string) string {
//
//	var sb strings.Builder
//	i := 1
//	for _, o := range options {
//		sb.WriteString("\n" + strconv.Itoa(i) + ". " + o)
//		i++
//	}
//	return sb.String()
//}
//
//func errorResponse(*Framework, *menu.Context, gateway.Gateway, *session.Session, string) gateway.Response {
//
//}
//
//func menuReplay() {
//
//}
//
//func menuOptionResponse() {
//
//}
//
//func newContext() *menu.Context {
//
//}
//
//func getOrCreateSession(gateway.Request) (*session.Session, error) {
//
//}
