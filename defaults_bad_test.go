package goconf

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// reflectValueOf builds a reflect.Value from one test literal.
func reflectValueOf(v SourceBytes) reflect.Value {
	return reflect.ValueOf(string(v))
}

// TestWithDefaultsRejectsNonStructValue: a default whose value is not
// struct-shaped cannot fill any target field and is refused at option
// construction, naming the value's shape.
func TestWithDefaultsRejectsNonStructValue(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("a non-struct default must be refused at option construction")
		}

		err, ok := r.(error)
		if !ok || !strings.Contains(err.Error(), "struct") {
			t.Errorf("refusal %v must be an error naming the required shape", r)
		}
	}()

	_ = WithDefaults("a string")
}

// TestLoadNeverPanicsOnBadDefault proves Load survives a bad default even
// if one is constructed: it surfaces a domain error naming WithDefaults.
func TestLoadNeverPanicsOnBadDefault(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
	})
	t.Chdir(filepath.Join(string(root), string(dirLocal)))

	bad := defaultsOption{value: reflectValueOf(SourceBytes("a string"))}

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(filepath.Join(string(root), string(dirLocal)))), &bad)
	if err == nil {
		t.Fatalf("a bad default must fail the load, not panic")
	}

	if !strings.Contains(err.Error(), "WithDefaults") {
		t.Errorf("error %q must name its source", err)
	}
}
