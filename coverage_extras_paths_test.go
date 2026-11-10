package goconf

import (
	"testing"
)

type (
	extraInt   int64
	extraInt8  int8
	extraUint  uint16
	extraFloat float64
	extraBool  bool
	extraSlice []extraVal
)

type intExtrasDoc struct {
	Extra map[extraKey]extraInt `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName               `json:"name"     toml:"name" yaml:"name"`
}

type int8ExtrasDoc struct {
	Extra map[extraKey]extraInt8 `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                `json:"name"     toml:"name" yaml:"name"`
}

type uintExtrasDoc struct {
	Extra map[extraKey]extraUint `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                `json:"name"     toml:"name" yaml:"name"`
}

type floatExtrasDoc struct {
	Extra map[extraKey]extraFloat `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                 `json:"name"     toml:"name" yaml:"name"`
}

type boolExtrasDoc struct {
	Extra map[extraKey]extraBool `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                `json:"name"     toml:"name" yaml:"name"`
}

type sliceExtrasDoc struct {
	Extra map[extraKey]extraSlice `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                 `json:"name"     toml:"name" yaml:"name"`
}

func TestExtrasTypedIntJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","n":42}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["n"] != 42 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedBoolJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","flag":true}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[boolExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedFloatJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","f":1.5}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[floatExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedUintJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","u":7}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[uintExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["u"] != 7 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedSliceJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","items":["a","b"]}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[sliceExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || len(got.Extra["items"]) != 2 || got.Extra["items"][0] != "a" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedIntYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nn: 42\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["n"] != 42 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedBoolYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nflag: true\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[boolExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedFloatYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nf: 1.5\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[floatExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedSliceYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nitems:\n  - a\n  - b\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[sliceExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || len(got.Extra["items"]) != 2 || got.Extra["items"][0] != "a" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
