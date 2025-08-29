package ussd

import (
	"fmt"
	u "github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/menu"
	"github.com/jamesdube/ussd/pkg/session"
)

type Normal struct {
}

func (n *Normal) Render(message string, ctx *menu.Context, sess *session.Session, f *Framework) menu.Response {

	if sess.Paginated {
		fmt.Println("pagination wanted")
	}

	sess.AddSelection(message)
	mn := f.router.RouteTo(sess.GetSelections())

	if mn == nil {

		u.Logger.Error("menu not found for route", "route", sess.GetSelections())
		return menu.Response{
			Prompt: "error",
		}
	}

	r := mn.OnRequest(ctx, message)

	if ctx.Paginated {
		sess.Paginated = true
		fmt.Println("create pagination")

		createPagination(ctx, r, sess)
	}

	f.SaveSession(sess)

	return menu.Response{
		Prompt:  r.Prompt,
		Options: r.Options,
	}
}

type Paginated struct {
}

func (n *Paginated) Render(message string, ctx *menu.Context, sess *session.Session, f *Framework) menu.Response {

	return menu.Response{
		Prompt: "paginated",
	}
}
