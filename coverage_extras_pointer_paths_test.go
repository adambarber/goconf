package goconf

import (
	"testing"
)

type ptrInner struct {
	Name resName `json:"name" toml:"name" yaml:"name"`
}

type ptrExtrasDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"     toml:"-"     yaml:"-"`
	Inner *ptrInner             `json:"inner"    toml:"inner" yaml:"inner"`
}

func TestExtrasPointerStructJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"inner":{"name":"n"},"x":"y"}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[ptrExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasPointerStructYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("inner:\n  name: n\nx: y\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[ptrExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasPointerStructTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("x = \"y\"\n[inner]\nname = \"n\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[ptrExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

type mapFieldDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-" toml:"-" yaml:"-"`
	M     map[extraKey]extraVal `json:"m"        toml:"m" yaml:"m"`
}

func TestExtrasMapFieldYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m:\n  a: b\nx: y\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[mapFieldDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.M["a"] != "b" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasMapFieldJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"m":{"a":"b"},"x":"y"}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[mapFieldDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.M["a"] != "b" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
