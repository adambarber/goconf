// Command extras_dynamic captures unmodelled keys into extras and accepts
// caller-chosen keys through a dynamic map.
package main

import (
	"fmt"

	"github.com/adambarber/goconf"
)

type (
	name      string
	extraKey  string
	extraWord string
	dynKey    string
	dynWord   string
)

type config struct {
	Extra map[extraKey]extraWord `goconf:"extras"  json:"-"   toml:"-"   yaml:"-"`
	Env   map[dynKey]dynWord     `goconf:"dynamic" json:"env" toml:"env" yaml:"env"`
	Name  name                   `json:"name"`
}

const appName goconf.Name = "app"

func main() {
	cfg, _, err := goconf.LoadBytes[config](
		[]byte(`{"name":"app","owner":"team","env":{"REGION":"eu"}}`),
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name, cfg.Extra["owner"], cfg.Env["REGION"])
}
