package goconf

import (
	"encoding/json/jsontext"
	"reflect"
	"strings"
	"testing"
)

func TestTOMLSetExtraValueNilCoverage(t *testing.T) {
	extras := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	key := reflect.ValueOf(extraKey("k"))

	if err := setExtraValue(extras, key, reflect.Value{}, FilePath("f.toml"), "k"); err != nil {
		t.Fatal(err)
	}
}

func TestTOMLSetExtraValueAssignableCoverage(t *testing.T) {
	extras := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraInt]()))
	key := reflect.ValueOf(extraKey("k"))
	val := reflect.ValueOf(extraInt(7))

	if err := setExtraValue(extras, key, val, FilePath("f.toml"), "k"); err != nil {
		t.Fatal(err)
	}
}

func TestTOMLExtrasTypeMismatchOnStructExtra(t *testing.T) {
	type inner struct {
		A extraVal `toml:"a"`
	}

	type doc struct {
		Extra map[extraKey]inner `goconf:"extras" toml:"-"`
		Name  resName            `toml:"name"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\n[x]\na = \"v\"\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestYAMLMapFieldScalarMismatch(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		M     map[extraKey]extraVal `yaml:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m: nope\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestYAMLMapFieldSequenceMismatch(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		M     map[extraKey]extraVal `yaml:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m:\n  - a\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestJSONExtrasNonMapExtrasFieldRefusal(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`1`), jsontext.AllowDuplicateNames(true))

	extras := reflect.New(reflect.TypeFor[int]()).Elem()
	if err := decodeJSONExtra(dec, SourceBytes(`{"x":1}`), FilePath("f.json"), extras, "", keySegment("x")); err == nil {
		t.Fatal("expected")
	}
}
