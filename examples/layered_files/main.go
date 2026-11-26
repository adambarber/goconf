// Command layered_files loads a global file first and a local file second;
// the local file overrides only the fields it names.
package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	appName  goconf.Name = "app"
	dirMode  os.FileMode = 0o700
	fileMode os.FileMode = 0o600
)

func discard(err error) {}

func main() {
	root, err := os.MkdirTemp("", "goconf-layers")
	if err != nil {
		panic(err)
	}
	defer func() { discard(os.RemoveAll(root)) }()

	global := filepath.Join(root, "global")
	local := filepath.Join(root, "local")

	for _, dir := range []string{global, local} {
		if mkdirErr := os.Mkdir(dir, dirMode); mkdirErr != nil {
			panic(mkdirErr)
		}
	}

	if writeErr := os.WriteFile(filepath.Join(global, "app.json"), []byte(`{"region":"global","port":1}`), fileMode); writeErr != nil {
		panic(writeErr)
	}

	if writeErr := os.WriteFile(filepath.Join(local, "app.json"), []byte(`{"region":"local"}`), fileMode); writeErr != nil {
		panic(writeErr)
	}

	cfg, _, err := goconf.Load[config](
		goconf.Spec{Name: appName, Format: goconf.FormatJSON},
		goconf.WithSearchPaths(goconf.DirPath(global), goconf.DirPath(local)),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
}
