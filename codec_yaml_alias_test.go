package goconf

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestYAMLAliasKeysArePresent pins that an alias resolves to its anchor's
// keys in the presence walk, so defaults never override them.
func TestYAMLAliasKeysArePresent(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("creds: &c\n  host: h\n  port: 1\nauth: *c\n")},
	})

	_, _, meta, err := LoadWithMeta[aliasCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	for _, p := range []OccurrencePath{"creds.host", "creds.port", "auth.host", "auth.port"} {
		if len(meta.ByPath(p)) != 1 {
			t.Errorf("alias key %q missing: %#v", p, meta.Entries())
		}
	}
}

// aliasCfg is the decode target of the alias tests.
type aliasCfg struct {
	Creds nestedCfg `json:"creds"`
	Auth  nestedCfg `json:"auth"`
}

// TestYAMLAliasDefaultDoesNotOverride pins the defaults consequence: a
// default layer filling auth.host must not override the alias-resolved
// value, because presence recorded auth.host.
func TestYAMLAliasDefaultDoesNotOverride(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("creds: &c\n  host: h\n  port: 1\nauth: *c\n")},
	})

	got, _, _, err := LoadWithMeta[aliasCfg](Spec{Name: configName("app"), Format: FormatYAML},
		append(loadOpts(globalPathOf(root)), WithDefaults(aliasCfg{
			Creds: nestedCfg{Host: "", Port: 0},
			Auth:  nestedCfg{Host: "DEFAULT", Port: 0},
		}))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if got.Auth.Host != "h" {
		t.Fatalf("default overrode an alias-resolved value: got %q", got.Auth.Host)
	}
}

// aliasCycleCfg holds a self-referencing alias without a recursive Go type:
// yaml.Node absorbs the alias node so the loader can still run end to end.
type aliasCycleCfg struct {
	Creds aliasCycleHost `yaml:"creds"`
}

type aliasCycleHost struct {
	Host mstr      `yaml:"host"`
	Self yaml.Node `yaml:"self"`
}

// TestYAMLAliasCycleDoesNotLoop pins that a self-referencing alias — legal
// yaml — terminates the public load path instead of looping.
func TestYAMLAliasCycleDoesNotLoop(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("creds: &c\n  host: h\n  self: *c\n")},
	})

	_, _, meta, err := LoadWithMeta[aliasCycleCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(meta.ByPath("creds.host")) == 0 {
		t.Fatalf("anchor keys must be present: %#v", meta.Entries())
	}
}

// TestYAMLScalarRootRefusalNamesLine pins that a scalar-root yaml
// document is refused with its line named.
func TestYAMLScalarRootRefusalNamesLine(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("# a comment\njust a scalar\n")},
	})

	_, _, err := Load[aliasCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err == nil {
		t.Fatalf("scalar root must be refused")
	}

	mustRefuseNames(t, err, "app.yaml:2:")
}
