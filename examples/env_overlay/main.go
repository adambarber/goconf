// Command env_overlay derives variable names from the schema and applies the
// environment as the outermost layer.
package main

import (
	"fmt"
	"os"

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

const appName goconf.Name = "app"

func discard(err error) {}

func main() {
	if err := os.Setenv("APP_PORT", "9090"); err != nil {
		panic(err)
	}

	defer func() { discard(os.Unsetenv("APP_PORT")) }()

	cfg, _, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithEnv("APP"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
}
