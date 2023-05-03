package m

import (
	"fmt"
	"github.com/jamesdube/ussd/pkg/menu"
)

type Foo struct {
}

func (h *Foo) OnRequest(ctx *menu.Context, msg string) menu.Response {
	ctx.NavigationType = menu.Stop
	fmt.Println("Foo onRequest")
	return menu.Response{
		Prompt: "foo",
	}
}

func (h *Foo) Process(ctx *menu.Context, msg string) menu.NavigationType {
	fmt.Println("Foo onProcess")
	return menu.Stop
}
