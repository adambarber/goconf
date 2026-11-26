// Command defaults fills only absent keys: a missing port gets the default,
// while an explicit null is a decision and stays zero.
package main

import (
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
	appName     goconf.Name = "app"
	defaultPort port        = 8080
)

func main() {
	missing, _, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu"}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithDefaults(config{Region: "", Port: defaultPort}),
	)
	if err != nil {
		panic(err)
	}

	explicitNull, _, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu","port":null}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithDefaults(config{Region: "", Port: defaultPort}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(missing.Region, missing.Port)
	fmt.Println(explicitNull.Region, explicitNull.Port)
}
