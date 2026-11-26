// Command secrets declares a secret-shaped field and renders a loaded config
// with that field redacted.
package main

import (
	"fmt"

	"github.com/adambarber/goconf"
)

type token string

type config struct {
	Token token `json:"token" secret:"true"`
}

const appName goconf.Name = "app"

func main() {
	cfg, _, err := goconf.LoadBytes[config](
		[]byte(`{"token":"hunter2"}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(goconf.Redact(cfg))
}
