package menu

type Menu interface {
	Header() string
	Render(c *Context, msg string) string
}

type Context struct {
	Context        map[string]string
	Msisdn         string
	NavigationType NavigationType
	Active         bool
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

func NewContext(msisdn string, attr map[string]string) *Context {
	return &Context{
		NavigationType: Continue,
		Context:        attr,
		Msisdn:         msisdn,
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
