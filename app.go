package main

import (
	"github.com/jamesdube/ussd/pkg/ussd"
	_ "github.com/jamesdube/ussd/pkg/ussd"
)

func main() {

	u := ussd.New()
	u.Start()

}
