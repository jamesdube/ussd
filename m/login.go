package m

import (
	"fmt"
	"github.com/jamesdube/ussd/pkg/menu"
)

type Login struct {
}

func (h *Login) OnRequest(ctx *menu.Context, msg string) menu.Response {

	if ctx.IsReplay() {
		return menu.Response{
			Prompt: "Please enter correct pin",
		}
	}

	return menu.Response{
		Prompt: "Please enter your pin",
	}

}

func (h *Login) Process(ctx *menu.Context, msg string) menu.NavigationType {
	b := true
	fmt.Println("login: ", b)

	if !b {
		ctx.NavigationType = menu.Replay
	}

	return menu.Continue
}
