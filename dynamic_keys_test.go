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
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"region":"eu","env":{"FOO":"1","BAR":"2"}}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[envCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Region != "eu" {
		t.Fatalf("region = %q", got.Region)
	}

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapKeysYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("region: eu\nenv:\n  FOO: \"1\"\n  BAR: \"2\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[envCfg](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}

func TestLoadDynamicMapKeysTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("region = \"eu\"\n[env]\nFOO = \"1\"\nBAR = \"2\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[envCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Env["FOO"] != "1" || got.Env["BAR"] != "2" {
		t.Fatalf("env = %#v", got.Env)
	}
}
