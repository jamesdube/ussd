package ussd

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	u "github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/session"
)

// requestState holds all state for a single USSD request.
// Methods on this struct replace the 7-8 parameter functions.
type requestState struct {
	framework *Framework
	gateway   gateway.Gateway
	request   gateway.Request
	session   *session.Session
	context   *menu.Context
}

// process returns a Fiber handler that processes USSD requests for the named gateway.
func process(framework *Framework, name string) func(ctx *fiber.Ctx) error {
	gw := framework.GetGateway(name)

	return func(ctx *fiber.Ctx) error {
		s, err := newRequestState(framework, gw, ctx)
		if err != nil {
			return err
		}
		return s.handle(ctx)
	}
}

// newRequestState parses the request, sets up the session, runs middleware,
// and returns a fully initialized requestState.
func newRequestState(f *Framework, gw gateway.Gateway, ctx *fiber.Ctx) (*requestState, error) {
	gr, err := gw.ToRequest(ctx)
	if err != nil {
		u.Logger.Error("failed to parse gateway request")
		return nil, ctx.Status(fiber.StatusBadRequest).SendString("failed to parse request")
	}

	ss, err := f.GetOrCreateSession(gr.SessionId)
	if err != nil {
		u.Logger.Error("failed to initiate session")
		r := buildResponse(gw, u.MenuInvalidSelection, nil, session.NewSession(gr.SessionId), gr.Msisdn, false)
		return nil, sendResponse(r, ctx)
	}

	err = runMiddleware(f, ss, gr)
	if err != nil {
		u.Logger.Error(err.Error())
		f.DeleteSession(ss.Id)
		r := buildResponse(gw, u.MenuInvalidSelection, nil, ss, gr.Msisdn, false)
		return nil, sendResponse(r, ctx)
	}

	c := menu.NewContext(gr.Msisdn, ss)

	return &requestState{
		framework: f,
		gateway:   gw,
		request:   gr,
		session:   ss,
		context:   c,
	}, nil
}

// handle is the top-level dispatch: resume pagination or process navigation.
func (s *requestState) handle(ctx *fiber.Ctx) error {
	if s.context.Paginated {
		return s.resumePagination(ctx)
	}
	return s.processNavigation(ctx)
}

// processNavigation routes to the current menu, calls Process() to determine
// navigation type, then dispatches to the appropriate handler.
func (s *requestState) processNavigation(ctx *fiber.Ctx) error {
	msg := s.request.Message

	currentMenu := s.framework.router.RouteTo(s.session.GetSelections())
	if currentMenu != nil {
		currentMenu.Process(s.context, msg)
	}

	switch s.context.NavigationType {
	case menu.Stop:
		return s.handleStop(ctx, currentMenu)
	case menu.Replay:
		return s.handleReplay(ctx, currentMenu)
	default:
		return s.handleContinue(ctx)
	}
}

// handleStop terminates the session and sends the final response.
func (s *requestState) handleStop(ctx *fiber.Ctx, currentMenu menu.Menu) error {
	if currentMenu == nil {
		return s.respondWithError(ctx)
	}

	response := currentMenu.OnRequest(s.context, s.request.Message)
	s.applyPostNavigation(response.NavigationType)
	return s.respond(ctx, response.Prompt, response.Options)
}

// handleReplay goes back to the previous menu and re-renders it.
func (s *requestState) handleReplay(ctx *fiber.Ctx, currentMenu menu.Menu) error {
	if currentMenu == nil {
		return s.respondWithError(ctx)
	}

	response := currentMenu.OnRequest(s.context, s.request.Message)
	s.applyPostNavigation(response.NavigationType)
	return s.respond(ctx, response.Prompt, response.Options)
}

// handleContinue advances to the next menu in the navigation tree.
func (s *requestState) handleContinue(ctx *fiber.Ctx) error {
	msg := s.request.Message

	s.session.AddSelection(msg)
	s.framework.SaveSession(s.session)

	nextMenu := s.framework.router.RouteTo(s.session.GetSelections())
	if nextMenu == nil {
		u.Logger.Error("menu not found for route", "route", s.session.GetSelections())
		return s.respondWithError(ctx)
	}

	response := nextMenu.OnRequest(s.context, msg)

	if response.Paginated {
		return s.handlePaginatedResponse(ctx, response)
	}

	s.applyPostNavigation(response.NavigationType)
	return s.respond(ctx, response.Prompt, response.Options)
}

// applyPostNavigation handles session lifecycle after navigation.
// Stop: delete session. Continue: save. Replay: go back two steps and save.
func (s *requestState) applyPostNavigation(navType menu.NavigationType) {
	switch navType {
	case menu.Stop:
		s.framework.DeleteSession(s.session.Id)
		s.context.Active = false
	case menu.Continue:
		s.framework.SaveSession(s.session)
	case menu.Replay:
		s.session.RemoveLastSelection()
		s.session.RemoveLastSelection()
		s.framework.SaveSession(s.session)
	}
}

// --- Pagination ---

// handlePaginatedResponse sets up pagination for a new paginated menu response.
func (s *requestState) handlePaginatedResponse(ctx *fiber.Ctx, response menu.Response) error {
	pages := paginateOptions(response.Options, response.PerPage)

	s.context.Pages = pages
	s.context.CurrentPage = 0
	s.session.Paginated = true
	s.session.PaginatedHasMore = len(pages) > 1
	s.session.Pages = pages
	s.session.CurrentPage = 0

	s.applyPostNavigation(response.NavigationType)
	return s.resumePagination(ctx)
}

// resumePagination handles an in-progress paginated menu.
// It decides whether to show the next page or process a selection.
func (s *requestState) resumePagination(ctx *fiber.Ctx) error {
	msg := s.request.Message
	isFirstView := s.session.CurrentPage == 0
	wantsNextPage := msg == "0"
	isLastPage := s.context.CurrentPage >= len(s.context.Pages)-1

	// Update hasMore flag when we're about to show the last page
	if s.context.CurrentPage+1 >= len(s.context.Pages) {
		s.session.PaginatedHasMore = false
	}

	// User pressed "More" on the last page — nothing left
	if isLastPage && wantsNextPage {
		return s.respondWithErrorMsg(ctx, u.MenuNoMoreOptions)
	}

	// User selected an option (not "More", not first view)
	userSelectedOption := !isFirstView && !wantsNextPage
	if userSelectedOption {
		return s.selectPaginatedOption(ctx, msg)
	}

	// Show next page (or first page)
	if !isFirstView && wantsNextPage {
		s.session.CurrentPage++
		s.context.CurrentPage = s.session.CurrentPage
		s.framework.SaveSession(s.session)
	}

	return s.respondWithPage(ctx, "Please select an option:", s.context.Pages[s.context.CurrentPage])
}

// selectPaginatedOption processes a user's selection from a paginated list.
func (s *requestState) selectPaginatedOption(ctx *fiber.Ctx, msg string) error {
	option, err := strconv.Atoi(msg)
	if err != nil || !s.isValidPageOption(option, s.context.CurrentPage) {
		u.Logger.Error("invalid pagination option", "route", s.session.GetSelections())
		return s.respondWithError(ctx)
	}

	s.context.SelectedPaginationOption = s.globalOptionIndex(option, s.context.CurrentPage)
	s.context.SelectedPageOption = option

	// Let the current menu process the selection
	currentMenu := s.framework.router.RouteTo(s.session.GetSelections())
	if currentMenu != nil {
		currentMenu.Process(s.context, msg)
	}

	// Navigate to the next menu
	s.session.AddSelection(msg)
	nextMenu := s.framework.router.RouteTo(s.session.GetSelections())
	if nextMenu == nil {
		u.Logger.Error("menu not found for route", "route", s.session.GetSelections())
		return s.respondWithError(ctx)
	}

	response := nextMenu.OnRequest(s.context, msg)
	s.applyPostNavigation(response.NavigationType)

	s.context.Paginated = false
	s.session.Paginated = false

	return s.respond(ctx, response.Prompt, response.Options)
}

// isValidPageOption checks if the option index is within bounds for the given page.
func (s *requestState) isValidPageOption(option int, pageIndex int) bool {
	if pageIndex < 0 || pageIndex >= len(s.context.Pages) {
		return false
	}

	page := s.context.Pages[pageIndex]
	return option >= 1 && option <= len(page)
}

// globalOptionIndex converts a local page option to a global option index
// across all pages viewed so far.
func (s *requestState) globalOptionIndex(localOption int, currentPage int) int {
	offset := 0
	for i := 0; i < currentPage; i++ {
		offset += len(s.context.Pages[i])
	}
	return localOption + offset
}

// paginateOptions splits a flat list of options into pages of the given size.
func paginateOptions(options []string, perPage int) [][]string {
	if perPage <= 0 || perPage >= len(options) {
		return [][]string{options}
	}

	var pages [][]string
	for i := 0; i < len(options); i += perPage {
		end := i + perPage
		if end > len(options) {
			end = len(options)
		}
		pages = append(pages, options[i:end])
	}
	return pages
}

// --- Response Helpers ---

// respond builds and sends a USSD response with the given prompt and options.
func (s *requestState) respond(ctx *fiber.Ctx, prompt string, options []string) error {
	r := buildResponse(s.gateway, prompt, options, s.session, s.request.Msisdn, s.context.Active)
	return sendResponse(r, ctx)
}

// respondWithPage builds and sends a paginated page response.
func (s *requestState) respondWithPage(ctx *fiber.Ctx, prompt string, pageOptions []string) error {
	r := buildResponse(s.gateway, prompt, pageOptions, s.session, s.request.Msisdn, s.context.Active)
	return sendResponse(r, ctx)
}

// respondWithError terminates the session and sends an invalid selection message.
func (s *requestState) respondWithError(ctx *fiber.Ctx) error {
	s.framework.DeleteSession(s.session.Id)
	r := buildResponse(s.gateway, u.MenuInvalidSelection, nil, s.session, s.request.Msisdn, false)
	return sendResponse(r, ctx)
}

// respondWithErrorMsg terminates the session and sends a custom error message.
func (s *requestState) respondWithErrorMsg(ctx *fiber.Ctx, msg string) error {
	u.Logger.Error(msg)
	s.framework.DeleteSession(s.session.Id)
	r := buildResponse(s.gateway, u.MenuInvalidSelection, nil, s.session, s.request.Msisdn, false)
	return sendResponse(r, ctx)
}

// --- Standalone Functions (signatures preserved for orchestrator.go compatibility) ---

// buildResponse constructs the gateway-specific response object.
func buildResponse(g gateway.Gateway, message string, options []string, session *session.Session, msisdn string, active bool) interface{} {
	m := message

	if len(options) > 0 {
		m += buildOptions(options)
	}

	if session.PaginatedHasMore && session.Paginated {
		m += "\n0. More"
	}

	return g.ToResponse(gateway.Response{
		Message:       m,
		Session:       session.GetID(),
		Msisdn:        msisdn,
		SessionActive: active,
	})
}

// buildOptions formats a slice of options as a numbered list.
func buildOptions(options []string) string {
	var sb strings.Builder
	for i, o := range options {
		sb.WriteString("\n" + strconv.Itoa(i+1) + ". " + o)
	}
	return sb.String()
}

// sendResponse marshals the response to XML and sends it.
func sendResponse(grs interface{}, ctx *fiber.Ctx) error {
	result, err := xml.Marshal(&grs)
	if err != nil {
		u.Logger.Error("failed to marshal XML response", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).SendString("internal error")
	}

	xmls := strings.ReplaceAll(string(result), "&#xA;", "\n")
	ctx.Type("xml")
	return ctx.Send([]byte(u.Header + xmls))
}

// runMiddleware executes the middleware chain.
func runMiddleware(f *Framework, ss *session.Session, gr gateway.Request) error {
	for _, m := range f.middlewareRegistry.Get() {
		if err := m.Handle(ss, &gr); err != nil {
			return err
		}
	}
	return nil
}
