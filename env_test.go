package goconf

import (
	"errors"
	"strings"
	"testing"
)

type envText string

type envOverlayCfg struct {
	Password secretText `json:"password" secret:"true" toml:"password" yaml:"password"`
	Port     int        `json:"port"     toml:"port"   yaml:"port"`
	Debug    bool       `json:"debug"    toml:"debug"  yaml:"debug"`
}

type envHostsCfg struct {
	Hosts []envText `json:"hosts" toml:"hosts" yaml:"hosts"`
}

type envAmbiguous struct {
	APIKey envText `json:"api_key" toml:"api_key" yaml:"api_key"`
	API    struct {
		Key envText `json:"key" toml:"key" yaml:"key"`
	} `json:"api" toml:"api" yaml:"api"`
}

func envSpec() Spec {
	return Spec{Name: configName("app"), Format: FormatTOML}
}

func TestEnvOverridesFileValue(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameTOML: SourceBytes("port = 8080\n"),
	})
	t.Setenv("APP_PORT", "9090")

	got, prov, err := Load[envOverlayCfg](envSpec(), WithSearchPaths(legPaths(root, dirLocal)...), WithEnv("APP"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Port != 9090 {
		t.Fatalf("env did not win: %+v", got)
	}

	if len(prov) != 1 {
		t.Fatalf("provenance entries = %d, want 1 (env is not a byte source)", len(prov))
	}
}

func TestEnvParsesScalarTypes(t *testing.T) {
	t.Setenv("APP_DEBUG", "true")

	got, _, err := Load[envOverlayCfg](envSpec(), WithEnv("APP"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !got.Debug {
		t.Fatalf("debug = %v, want true", got.Debug)
	}
}

func TestEnvRefusesBadValueWithoutEcho(t *testing.T) {
	planted := "zzq-secret-marker-8842"
	t.Setenv("APP_PORT", planted)

	_, _, err := Load[envOverlayCfg](envSpec(), WithEnv("APP"))
	if !errors.Is(err, ErrEnvParse) {
		t.Fatalf("err = %v, want ErrEnvParse", err)
	}

	if strings.Contains(err.Error(), planted) {
		t.Fatalf("env value leaked: %s", err.Error())
	}

	want := "Error: APP_PORT:1: invalid environment value: APP_PORT. set it to a whole number."
	if err.Error() != want {
		t.Fatalf("render = %q, want %q", err.Error(), want)
	}
}

func TestEnvAppliesSecretGate(t *testing.T) {
	t.Setenv("APP_PASSWORD", "hunter2")

	_, _, err := Load[unmarkedCfg](envSpec(), WithEnv("APP"))
	if !errors.Is(err, ErrNeedDeclaration) {
		t.Fatalf("err = %v, want ErrNeedDeclaration", err)
	}
}

func TestEnvAmbiguousNameRefused(t *testing.T) {
	t.Setenv("APP_API_KEY", "value")

	_, _, err := Load[envAmbiguous](envSpec(), WithEnv("APP"))
	if !errors.Is(err, ErrEnvAmbiguous) {
		t.Fatalf("err = %v, want ErrEnvAmbiguous", err)
	}
}

func TestEnvCompositeJSON(t *testing.T) {
	t.Setenv("APP_HOSTS", `["a","b"]`)

	got, _, err := Load[envHostsCfg](envSpec(), WithEnv("APP"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(got.Hosts) != 2 || got.Hosts[0] != "a" {
		t.Fatalf("hosts = %v", got.Hosts)
	}
}

func TestEnvUnknownVariableIgnored(t *testing.T) {
	t.Setenv("APP_NOT_IN_SCHEMA", "whatever")

	if _, _, err := Load[envOverlayCfg](envSpec(), WithEnv("APP")); err != nil {
		t.Fatalf("unmapped variable must be ignored: %v", err)
	}
}

func FuzzEnvValueFreedom(f *testing.F) {
	f.Add([]byte("zzq-secret-marker-8842"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, value []byte) {
		// NUL cannot be set in a process environment; fold it so the fuzz
		// target stays reachable for every byte string.
		planted := "ZZQ-" + strings.ReplaceAll(string(value), "\x00", "x")
		t.Setenv("APP_PORT", planted)

		_, _, err := Load[envOverlayCfg](envSpec(), WithEnv("APP"))
		if err == nil {
			return
		}

		if strings.Contains(err.Error(), planted) {
			t.Fatalf("env value leaked: %s", err.Error())
		}
	})
}
