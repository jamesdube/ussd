package m

import (
	"errors"
	"fmt"
	"github.com/jamesdube/ussd/pkg/gateway"
	"github.com/jamesdube/ussd/pkg/session"
)

type Auth struct {
}

func (a *Auth) Handle(s *session.Session, gr *gateway.Request) error {

	fmt.Println("auth---->")
	return errors.New("need token")

}
