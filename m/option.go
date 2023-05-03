package m

import (
	"fmt"
	"github.com/jamesdube/ussd/pkg/menu"
)

type Option struct {
}

func (h *Option) OnRequest(ctx *menu.Context, msg string) menu.Response {

	fmt.Println("msg:", msg)

	ctx.NavigationType = menu.Stop
	return menu.Response{
		Prompt: "you selected: " + ctx.Get(selected),
	}
}

func (h *Option) Process(ctx *menu.Context, msg string) menu.NavigationType {
	fmt.Println("Paginated on Process: ", ctx.SelectedPaginationOption, " selected")

	return menu.Stop
}
