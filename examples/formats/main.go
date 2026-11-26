// Command formats shows that JSON, YAML, and TOML share one schema and one
// strict decode contract.
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
	yamlCfg, _, err := goconf.LoadBytes[config](
		[]byte("region: eu\nport: 8080\n"),
		goconf.Spec{Name: appName, Format: goconf.FormatYAML},
	)
	if err != nil {
		panic(err)
	}

	tomlCfg, _, err := goconf.LoadBytes[config](
		[]byte("region = \"eu\"\nport = 8080\n"),
		goconf.Spec{Name: appName, Format: goconf.FormatTOML},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(yamlCfg.Region, yamlCfg.Port)
	fmt.Println(tomlCfg.Region, tomlCfg.Port)
}
