package m

import (
	"fmt"
	"github.com/jamesdube/ussd/pkg/menu"
)

type Home struct {
}

func (h *Home) OnRequest(ctx *menu.Context, msg string) menu.Response {

	return menu.Response{
		Prompt:  "home screen",
		Options: []string{"foo", "bar"},
	}
}

func (h *Home) Process(ctx *menu.Context, msg string) menu.NavigationType {
	fmt.Println("home")
	return menu.Continue
}
