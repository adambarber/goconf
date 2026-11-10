package goconf

import (
	"testing"
)

type extraStruct struct {
	X resName `json:"x" toml:"x" yaml:"x"`
}

type structExtrasDoc struct {
	Extra map[extraKey]extraStruct `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                  `json:"name"     toml:"name" yaml:"name"`
}

type float32ExtrasDoc struct {
	Extra map[extraKey]float32Val `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName                 `json:"name"     toml:"name" yaml:"name"`
}

type float32Val float32

func TestExtrasStructValueJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","s":{"x":"v"}}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[structExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["s"].X != "v" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasStructValueYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\ns:\n  x: v\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[structExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["s"].X != "v" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasStructValueTOMLRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\n[s]\nx = \"v\"\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[structExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestExtrasFloat32TOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nf = 1.5\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[float32ExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasJSONArrayInTypedSliceJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"name":"r","items":["a","b","c"]}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[sliceExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || len(got.Extra["items"]) != 3 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasYAMLMapValue(t *testing.T) {
	type mapVal map[extraKey]extraVal

	type mapExtrasDoc struct {
		Extra map[extraKey]mapVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
		Name  resName             `json:"name"     toml:"name" yaml:"name"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nm:\n  a: b\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[mapExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["m"]["a"] != "b" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
