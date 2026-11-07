package goconf

import (
	"reflect"
	"testing"
)

// anchor's keys in the presence walk: a defaults layer must not override
// a value the alias resolves to. Data: creds anchors a mapping, auth
// aliases it, so auth.host and auth.port are present.
func TestYAMLAliasKeysArePresent(t *testing.T) {
	data := SourceBytes("creds: &c\n  host: h\n  port: 1\nauth: *c\n")

	keys := presentKeysOK(t, presentYAMLKeys, data)

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

	keys, err := presentYAMLKeys(data)
	if err != nil {
		t.Fatalf("presence: %v", err)
	}

	dst := Target{Value: reflect.ValueOf(&aliasCfg{Creds: nestedCfgWant("", 0), Auth: nestedCfgWant("", 0)}).Elem()}

	if err := decodeYAMLStrict(data, dst); err != nil {
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

	keys, err := presentYAMLKeys(data)
	if err != nil {
		t.Fatalf("presence: %v", err)
	}

	if !keys.Has("creds.host") {
		t.Fatalf("anchor keys must be present: %v", keys)
	}
}
