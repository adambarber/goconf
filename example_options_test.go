package goconf_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/adambarber/gorefuse"

	"github.com/adambarber/goconf"
)

func ExampleWithDefaults() {
	cfg, _, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu"}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
		goconf.WithDefaults(exampleConfig{Region: "", Port: 8080}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleWithEnv() {
	if err := os.Setenv("APP_PORT", "9090"); err != nil {
		panic(err)
	}

	defer func() { ignoreCleanupErr(os.Unsetenv("APP_PORT")) }()

	cfg, _, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
		goconf.WithEnv("APP"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port)
	// Output: eu 9090
}

func ExampleWithMaxBytes() {
	_, _, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu","port":8080}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
		goconf.WithMaxBytes(8),
	)

	fmt.Println(errors.Is(err, goconf.ErrSourceTooLarge))
	// Output: true
}

func ExampleReason() {
	_, _, err := goconf.LoadBytes[exampleConfig](
		[]byte(`{"region":"eu","typo":1}`),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatJSON},
	)

	because, ok := errors.AsType[gorefuse.Because[goconf.Reason]](err)
	fmt.Println(ok, because.Reason == goconf.ErrUnknownField)
	// Output: true true
}

type exampleValueDoc struct {
	Count goconf.Value `toml:"count"`
}

func ExampleDecodeValue() {
	cfg, _, err := goconf.LoadBytes[exampleValueDoc](
		[]byte("count = 42\n"),
		goconf.Spec{Name: exampleAppName, Format: goconf.FormatTOML},
	)
	if err != nil {
		panic(err)
	}

	count, err := goconf.DecodeValue[int](cfg.Count)
	if err != nil {
		panic(err)
	}

	fmt.Println(count)
	// Output: 42
}
