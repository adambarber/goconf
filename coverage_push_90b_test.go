package goconf

import (
	"errors"
	"testing"
)

func TestTOMLFoundTermMoreCoverage(t *testing.T) {
	_ = tomlFoundTerm(tomlMessage("expected string but found bool"))
	_ = tomlFoundTerm(tomlMessage("expected integer but found string"))
	_ = tomlFoundTerm(tomlMessage("expected float but found string"))
	_ = tomlFoundTerm(tomlMessage("expected array but found table"))
	_ = tomlFoundTerm(tomlMessage("something else entirely"))
}

func TestYAMLErrValueFreeTypeError(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		N     int                   `yaml:"n"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("n: [1,2]\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err=%v", err)
	}
}

func TestJSONExtrasRootShapeArrayIntoMap(t *testing.T) {
	type m map[extraKey]extraVal

	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
		M     m                     `json:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"m":[1,2]}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestYAMLMapNodeDefaultBranch(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		M     map[extraKey]extraVal `yaml:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m: !!str nope\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}
