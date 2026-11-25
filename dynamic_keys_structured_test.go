package goconf

import "testing"

type (
	nestedKey  string
	nestedHost string
)

type nestedDynCfg struct {
	Env map[nestedKey]nestedEnv `goconf:"dynamic" json:"env" toml:"env" yaml:"env"`
}

type nestedEnv struct {
	Host nestedHost `json:"host" toml:"host" yaml:"host"`
	Port int        `json:"port" toml:"port" yaml:"port"`
}

func TestLoadDynamicMapStructuredValuesJSON(t *testing.T) {
	got := loadBytesOK[nestedDynCfg](t, SourceBytes(`{"env":{"a":{"host":"h","port":1},"b":{"host":"x","port":2}}}`), FormatJSON)

	if got.Env["a"].Host != "h" || got.Env["a"].Port != 1 || got.Env["b"].Port != 2 {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapStructuredValuesYAML(t *testing.T) {
	got := loadBytesOK[nestedDynCfg](t, SourceBytes("env:\n  a:\n    host: h\n    port: 1\n  b:\n    host: x\n    port: 2\n"), FormatYAML)

	if got.Env["a"].Host != "h" || got.Env["b"].Port != 2 {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapStructuredValuesTOML(t *testing.T) {
	got := loadBytesOK[nestedDynCfg](t, SourceBytes("[env.a]\nhost = \"h\"\nport = 1\n[env.b]\nhost = \"x\"\nport = 2\n"), FormatTOML)

	if got.Env["a"].Host != "h" || got.Env["b"].Port != 2 {
		t.Fatalf("env = %#v", got.Env)
	}
}
