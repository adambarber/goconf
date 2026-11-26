// Command value holds an opaque TOML value and decodes it into the type your
// application wants.
package main

import (
	"fmt"

	"github.com/adambarber/goconf"
)

type countDoc struct {
	Count goconf.Value `toml:"count"`
}

const appName goconf.Name = "app"

func main() {
	cfg, _, err := goconf.LoadBytes[countDoc](
		[]byte("count = 42\n"),
		goconf.Spec{Name: appName, Format: goconf.FormatTOML},
	)
	if err != nil {
		panic(err)
	}

	count, err := goconf.DecodeValue[int](cfg.Count)
	if err != nil {
		panic(err)
	}

	fmt.Println(count)
}
