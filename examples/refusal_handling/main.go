// Command refusal_handling extracts the typed reason and frame from a refusal.
package main

import (
	"errors"
	"fmt"

	"github.com/adambarber/gorefuse"

	"github.com/adambarber/goconf"
)

type region string

type config struct {
	Region region `json:"region"`
}

const appName goconf.Name = "app"

func main() {
	_, _, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu","typo":1}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
	)

	because, ok := errors.AsType[gorefuse.Because[goconf.Reason]](err)
	fmt.Println(ok, because.Reason == goconf.ErrUnknownField, because.What.File, because.What.Line)
}
