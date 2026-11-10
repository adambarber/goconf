package goconf

import (
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestTOMLFoundTermVariants(t *testing.T) {
	_ = tomlFoundTerm(tomlMessage("expected integer but found float"))
	_ = tomlFoundTerm(tomlMessage("expected bool but found string"))
	_ = tomlFoundTerm(tomlMessage("expected table but found array"))
	_ = tomlFoundTerm(tomlMessage("found string"))
}

func TestDerefValueNilPointer(t *testing.T) {
	var p *intExtrasDoc

	_ = derefValue(reflect.ValueOf(p))
}

func TestYAMLPopulateMapNodeSequenceNested(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		M     map[extraKey]extraVal `yaml:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m: [1, 2]\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestTOMLCaptureKeyInvalidDest(t *testing.T) {
	raw := reflect.ValueOf(map[string]int{"a": 1})

	ok, err := tomlCaptureKey(reflect.Value{}, raw, toml.Key{"a"}, FilePath("f.toml"))
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestTOMLSetExtraInterfaceElem(t *testing.T) {
	extras := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[interface{ Error() string }]()))
	key := reflect.ValueOf(extraKey("k"))
	val := reflect.ValueOf("x")

	if err := setExtraValue(extras, key, val, FilePath("f.toml"), "k"); err == nil {
		t.Fatal("expected refusal")
	}
}
