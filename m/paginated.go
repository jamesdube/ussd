package m

import (
	"fmt"
	"github.com/jamesdube/ussd/pkg/menu"
)

const selected = "SELECTED_OPTION"

type Paginated struct {
}

func (h *Paginated) OnRequest(ctx *menu.Context, msg string) menu.Response {

	ctx.NavigationType = menu.Continue
	ctx.Paginated = true

	return menu.Response{
		Prompt:  "foo",
		Options: []string{"one", "two", "three", "four"},
		PerPage: 2,
	}
}

func (h *Paginated) Process(ctx *menu.Context, msg string) menu.NavigationType {

	val := ctx.Pages[ctx.CurrentPage-1][ctx.SelectedPaginationOption-1]

	fmt.Println("Paginated on Process:", msg, "selected", "value:", val)

	ctx.Add(selected, val)

	return menu.Stop
}
