package goconf

import (
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestTOMLCaptureKeyMissing(t *testing.T) {
	ok, err := tomlCaptureKey(reflect.Value{}, reflect.ValueOf(map[string]int{"a": 1}), toml.Key{"b"}, tomlAOTIndex{}, FilePath("f.toml"))
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
