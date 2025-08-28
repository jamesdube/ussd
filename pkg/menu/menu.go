package menu

import "github.com/jamesdube/ussd/pkg/session"

type Menu interface {
	OnRequest(c *Context, msg string) Response
	Process(ctx *Context, msg string) NavigationType
}

type Response struct {
	Prompt         string
	Options        []string
	Paginated      bool
	PerPage        int
	NavigationType NavigationType
}

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

type Registry struct {
	menus map[string]Menu
}

func (r *Registry) Add(n string, m Menu) {
	r.menus[n] = m
}

func (r *Registry) Find(m string) Menu {
	mn := r.menus[m]
	return mn
}

func NewRegistry() *Registry {
	return &Registry{menus: map[string]Menu{}}
}

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

/*type ResponseRenderer interface {
	Render() gateway.Response
}

func NewMenu(n string, k string) *Menu {
	return &Menu{
		name:     n,
		routeKey: k,
	}
}

func (m *Menu) GetRouteKey() string {
	return m.routeKey
}

func (m *Menu) GetName() string {
	return m.name
}

func (m *Menu) Render() gateway.Response {
	return gateway.Response{}
}*/
