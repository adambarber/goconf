package goconf

import (
	"errors"
	"testing"
)

type (
	overflowInt int8
	overflowDoc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
		N     overflowInt           `json:"n"`
	}
)

func TestExtrasJSONIntOverflowRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"n":99999,"x":"y"}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[overflowDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err = %v, want ErrTypeMismatch", err)
	}
}

type fixedArrDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
	Arr   [2]int                `json:"arr"`
}

func TestExtrasJSONFixedArrayOverflowRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"arr":[1,2,3],"x":"y"}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[fixedArrDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestExtrasJSONRootShapeRefusedAsNotAMapping(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`[1,2,3]`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrNotAMapping) {
		t.Fatalf("err = %v, want ErrNotAMapping", err)
	}
}

type yamlIntExtraDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
	N     int                   `yaml:"n"`
}

func TestExtrasYAMLTypeMismatchRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("n: hello\nx: y\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[yamlIntExtraDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err = %v, want ErrTypeMismatch", err)
	}
}
