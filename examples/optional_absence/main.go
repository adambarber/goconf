// Command optional_absence lets a fresh install start from defaults when no
// config file exists, while still refusing real IO faults.
package main

import (
	"fmt"
	"log"
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

const (
	appName     goconf.Name = "app"
	defaultPort port        = 8080
)

func discard(err error) {}

func main() {
	dir, err := os.MkdirTemp("", "goconf-absent")
	if err != nil {
		panic(err)
	}
	defer func() { discard(os.RemoveAll(dir)) }()

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	cfg, _, err := goconf.Load[config](
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithSearchPaths(goconf.DirPath(dir)),
		goconf.WithOptionalConfig(),
		goconf.WithDefaults(config{Region: "eu", Port: defaultPort}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
}
