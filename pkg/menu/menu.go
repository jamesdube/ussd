package menu

import "github.com/jamesdube/ussd/pkg/session"

// NavigationType controls the flow after a menu processes a request.
type NavigationType int64

const (
	Continue  NavigationType = 0
	Stop      NavigationType = 1
	Replay    NavigationType = 2
	Paginated NavigationType = 3
	LongCode  NavigationType = 4
)

// Menu defines the interface for USSD menu screens.
type Menu interface {
	OnRequest(c *Context, msg string) Response
	Process(ctx *Context, msg string) NavigationType
}

// Response is what a menu returns after processing a request.
type Response struct {
	Prompt         string
	Options        []string
	Paginated      bool
	PerPage        int
	NavigationType NavigationType
}

// Context carries request-scoped state through the menu processing pipeline.
type Context struct {
	Context                  map[string]string
	Msisdn                   string
	NavigationType           NavigationType
	Paginated                bool
	Pages                    [][]string
	CurrentPage              int
	SelectedPaginationOption int
	SelectedPageOption       int
	Active                   bool
}

func (d *Context) Add(k string, v string) {
	d.Context[k] = v
}

func (d *Context) Get(k string) string {
	return d.Context[k]
}

func (d *Context) GetData() map[string]string {
	return d.Context
}

func (d *Context) IsReplay() bool {
	return d.NavigationType == Replay
}

// Registry stores named menu implementations.
type Registry struct {
	menus map[string]Menu
}

func (r *Registry) Add(n string, m Menu) {
	r.menus[n] = m
}

func (r *Registry) Find(m string) Menu {
	return r.menus[m]
}

func NewRegistry() *Registry {
	return &Registry{menus: map[string]Menu{}}
}

// NewContext creates a menu context from the current session state.
func NewContext(msisdn string, session *session.Session) *Context {
	return &Context{
		NavigationType: Continue,
		Context:        session.Attributes,
		Msisdn:         msisdn,
		Active:         true,
		Paginated:      session.Paginated,
		Pages:          session.Pages,
		CurrentPage:    session.CurrentPage,
	}
}
