package goconf

import "testing"

type (
	dynKey    string
	dynVal    string
	dynRegion string
)

type envCfg struct {
	Env    map[dynKey]dynVal `goconf:"dynamic" json:"env"    toml:"env"    yaml:"env"`
	Region dynRegion         `json:"region"    toml:"region" yaml:"region"`
}

func TestLoadDynamicMapKeysJSON(t *testing.T) {
	got := loadBytesOK[envCfg](t, SourceBytes(`{"region":"eu","env":{"FOO":"1","BAR":"2"}}`), FormatJSON)

	if got.Region != "eu" {
		t.Fatalf("region = %q", got.Region)
	}

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapKeysYAML(t *testing.T) {
	got := loadBytesOK[envCfg](t, SourceBytes("region: eu\nenv:\n  FOO: \"1\"\n  BAR: \"2\"\n"), FormatYAML)

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapKeysTOML(t *testing.T) {
	got := loadBytesOK[envCfg](t, SourceBytes("region = \"eu\"\n[env]\nFOO = \"1\"\nBAR = \"2\"\n"), FormatTOML)

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}
