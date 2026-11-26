// Command max_bytes caps one source before decode, so an oversized document
// is refused without being parsed.
package main

import (
	"errors"
	"fmt"

	"github.com/adambarber/goconf"
)

type (
	region string
	port   int
)

type config struct {
	Region region `json:"region"`
	Port   port   `json:"port"`
}

const (
	appName   goconf.Name = "app"
	byteLimit             = 8
)

func main() {
	_, _, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithMaxBytes(byteLimit),
	)

	fmt.Println(errors.Is(err, goconf.ErrSourceTooLarge))
}
