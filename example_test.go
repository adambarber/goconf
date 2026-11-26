package goconf_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adambarber/goconf"
)

type exampleRegion string

const exampleAppName goconf.Name = "app"

// ignoreCleanupErr keeps example cleanup explicit without adding output.
func ignoreCleanupErr(err error) {}

type exampleConfig struct {
	Region exampleRegion `json:"region" toml:"region" yaml:"region"`
	Port   int           `json:"port"   toml:"port"   yaml:"port"`
}

func Example() {
	cfg, prov, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port, len(prov))
	// Output: eu 8080 1
}

func ExampleLoadBytes() {
	cfg, _, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu"}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
		goconf.WithDefaults(exampleConfig{Region: "", Port: 8080}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
	// Output: eu 8080
}

func ExampleLoad() {
	dir, err := os.MkdirTemp("", "goconf-example")
	if err != nil {
		panic(err)
	}
	defer func() { ignoreCleanupErr(os.RemoveAll(dir)) }()

	file := filepath.Join(dir, "app.json")
	if writeErr := os.WriteFile(file, []byte(`{"region":"eu","port":8080}`), 0o600); writeErr != nil {
		panic(writeErr)
	}

	cfg, _, err := goconf.Load[exampleConfig](
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
		goconf.WithSearchPaths(goconf.DirPath(dir)),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
	// Output: eu 8080
}

func ExampleLoadWithMeta() {
	_, _, meta, err := goconf.LoadBytesWithMeta[exampleConfig](
		[]byte("{\n  \"region\": \"eu\",\n  \"port\": 8080\n}"),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	occurrences := meta.ByPath(goconf.OccurrencePath("port"))
	fmt.Println(occurrences[0].Path, occurrences[0].Line, occurrences[0].Effective,
		occurrences[0].LineConfidence == goconf.LineExact)
	// Output: port 3 true true
}

type exampleToken string

type exampleSecretConfig struct {
	Token exampleToken `json:"token" secret:"true"`
}

func ExampleRedact() {
	cfg, _, err := goconf.LoadBytes[exampleSecretConfig](
		[]byte(`{"token":"hunter2"}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(goconf.Redact(cfg))
	// Output: {Token:[redacted]}
}
