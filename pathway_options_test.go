package goconf

import (
	"errors"
	"strings"
	"testing"
)

type pathwayToken string

type (
	pathwayExtraKey  string
	pathwayExtraWord string
	pathwayDynKey    string
	pathwayDynWord   string
)

type pathwaySecretConfig struct {
	Token pathwayToken `json:"token" secret:"true"`
}

type pathwayExtrasConfig struct {
	Extra map[pathwayExtraKey]pathwayExtraWord `goconf:"extras"  json:"-"   toml:"-"   yaml:"-"`
	Env   map[pathwayDynKey]pathwayDynWord     `goconf:"dynamic" json:"env" toml:"env" yaml:"env"`
	Name  pathwayRegion                        `json:"name"`
}

type pathwayValueDoc struct {
	Count Value `toml:"count"`
}

const (
	pathwayEnvPort     pathwayPort = 9090
	pathwayEnvPortText             = "9090"
	pathwayCount                   = 42
)

func TestPublicPathwayEnvOverlay(t *testing.T) {
	t.Setenv("APP_PORT", pathwayEnvPortText)

	cfg, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu","port":8080}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithEnv("APP"),
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Port != pathwayEnvPort {
		t.Fatalf("port = %d, want %d", cfg.Port, pathwayEnvPort)
	}
}

func TestPublicPathwayOptionalAbsence(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{})

	cfg, _, err := Load[pathwayConfig](
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...),
		WithOptionalConfig(),
		WithDefaults(pathwayConfig{Region: "eu", Port: pathwayPortFull}),
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Region != "eu" || cfg.Port != pathwayPortFull {
		t.Fatalf("cfg = %+v, want defaults", cfg)
	}
}

func TestPublicPathwayMaxBytes(t *testing.T) {
	_, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu","port":8080}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithMaxBytes(pathwayByteCap),
	)

	if !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v, want ErrSourceTooLarge", err)
	}
}

func TestPublicPathwaySecrets(t *testing.T) {
	cfg, _, err := LoadBytes[pathwaySecretConfig](
		[]byte(`{"token":"hunter2"}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	rendered := string(Redact(cfg))
	if !strings.Contains(rendered, "[redacted]") || strings.Contains(rendered, "hunter2") {
		t.Fatalf("rendered = %q, want token redacted", rendered)
	}
}

func TestPublicPathwayExtrasAndDynamic(t *testing.T) {
	cfg, _, err := LoadBytes[pathwayExtrasConfig](
		[]byte(`{"name":"app","owner":"team","env":{"REGION":"eu"}}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Extra["owner"] != "team" || cfg.Env["REGION"] != "eu" {
		t.Fatalf("extras = %+v, dynamic = %+v", cfg.Extra, cfg.Env)
	}
}

func TestPublicPathwayValue(t *testing.T) {
	cfg, _, err := LoadBytes[pathwayValueDoc](
		[]byte("count = 42\n"),
		Spec{Name: pathwayAppName, Format: FormatTOML},
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	count, err := DecodeValue[int](cfg.Count)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if count != pathwayCount {
		t.Fatalf("count = %d, want %d", count, pathwayCount)
	}
}
