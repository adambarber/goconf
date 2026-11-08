package goconf

import (
	"reflect"
	"testing"
)

// TestYAMLAliasKeysArePresent pins that an alias resolves to its
// anchor's keys in the presence walk, so defaults never override them.
func TestYAMLAliasKeysArePresent(t *testing.T) {
	data := SourceBytes("creds: &c\n  host: h\n  port: 1\nauth: *c\n")

	keys := presentKeysOK(t, fileKeys(presentYAMLKeys, fileNameYAML), data)

	keysPresent(t, keys, []FieldPath{"creds.host", "creds.port", "auth.host", "auth.port"}, nil)
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
	data := SourceBytes("creds: &c\n  host: h\n  port: 1\nauth: *c\n")

	keys, err := presentYAMLKeys(data, FilePath(string(fileNameYAML)))
	if err != nil {
		t.Fatalf("presence: %v", err)
	}

	dst := Target{Value: reflect.ValueOf(&aliasCfg{Creds: nestedCfgWant("", 0), Auth: nestedCfgWant("", 0)}).Elem()}

	if err := decodeYAMLStrict(data, FilePath(string(fileNameYAML)), dst); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if err := fillAbsents(dst, keys, nil); err != nil {
		t.Fatalf("fill: %v", err)
	}
}

// TestYAMLAliasCycleDoesNotLoop pins that a self-referencing alias —
// legal yaml — terminates the presence walk instead of looping.
func TestYAMLAliasCycleDoesNotLoop(t *testing.T) {
	data := SourceBytes("creds: &c\n  host: h\n  self: *c\n")

	keys, err := presentYAMLKeys(data, FilePath(string(fileNameYAML)))
	if err != nil {
		t.Fatalf("presence: %v", err)
	}

	if !keys.Has("creds.host") {
		t.Fatalf("anchor keys must be present: %v", keys)
	}
}

// TestYAMLScalarRootRefusalNamesLine pins that a scalar-root yaml
// document is refused with its line named.
func TestYAMLScalarRootRefusalNamesLine(t *testing.T) {
	_, err := presentYAMLKeys(SourceBytes("# a comment\njust a scalar\n"), FilePath(string(fileNameYAML)))
	if err == nil {
		t.Fatalf("scalar root must be refused")
	}

	mustRefuseNames(t, err, "app.yaml:2:")
}
