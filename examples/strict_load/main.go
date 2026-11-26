// Command strict_load decodes bytes into a typed struct and reports the exact
// source digest. Copy this file and replace the config shape.
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
	Region region `json:"region" toml:"region" yaml:"region"`
	Port   port   `json:"port"   toml:"port"   yaml:"port"`
}

const appName goconf.Name = "app"

func main() {
	cfg, prov, err := goconf.LoadBytes[config](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port, len(prov))
}
