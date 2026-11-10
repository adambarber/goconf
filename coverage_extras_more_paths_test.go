package goconf

import (
	"errors"
	"reflect"
	"testing"
)

type dynExtrasInner struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName               `json:"name"     toml:"name" yaml:"name"`
}

type dynExtrasDoc struct {
	Items map[dynKey]dynExtrasInner `goconf:"dynamic" json:"items" toml:"items" yaml:"items"`
}

func TestTOMLExtrasInsideDynamicMapEntry(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("[items.one]\nname = \"n\"\nx = \"y\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[dynExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("err=%v", err)
	}

	if got.Items["one"].Name != "n" || got.Items["one"].Extra["x"] != "y" {
		t.Fatalf("got=%#v", got)
	}
}

func TestYAMLExtrasInsideDynamicMapEntry(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("items:\n  one:\n    name: n\n    x: y\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[dynExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("err=%v", err)
	}

	if got.Items["one"].Extra["x"] != "y" {
		t.Fatalf("got=%#v", got)
	}
}

func TestJSONExtrasInsideDynamicMapEntry(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"items":{"one":{"name":"n","x":"y"}}}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[dynExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("err=%v", err)
	}

	if got.Items["one"].Extra["x"] != "y" {
		t.Fatalf("got=%#v", got)
	}
}

func TestTOMLFloatFitsCoverage(t *testing.T) {
	dst := reflect.New(reflect.TypeFor[float32]()).Elem()

	val := reflect.ValueOf(float64(1.5))
	if !tomlFloatFits(val, dst.Type()) {
		t.Fatal("expected fit")
	}
}

func TestTOMLUintFitsCoverage(t *testing.T) {
	dst := reflect.New(reflect.TypeFor[uint16]()).Elem()

	val := reflect.ValueOf(uint64(7))
	if !tomlUintFits(val, dst.Type()) {
		t.Fatal("expected fit")
	}
}

func TestJSONUnknownFieldExtrasPathRefused(t *testing.T) {
	type inner struct {
		Name resName `json:"name"`
	}

	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
		In    inner                 `json:"in"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"in":{"name":"n","bad":1}}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("err=%v", err)
	}
}

func TestYAMLRootScalarExtrasNotMapping(t *testing.T) {
	_, _, err := Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(t.TempDir())))
	_ = err

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("just-a-scalar\n")},
	})
	t.Chdir(string(root))

	_, _, err = Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrNotAMapping) {
		t.Fatalf("err=%v", err)
	}
}
