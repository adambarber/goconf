package goconf

import "testing"

type caseResName string

type caseExtrasDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-" toml:"-" yaml:"-"`
	Name  caseResName           `json:"name"`
}

func TestExtrasJSONCaseInsensitiveFixedField(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"NAME":"r","provider":"aws"}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[caseExtrasDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Name != "r" {
		t.Fatalf("Name = %q, want r", got.Name)
	}

	if got.Extra[extraKey("provider")] != extraVal("aws") {
		t.Fatalf("extra provider = %#v", got.Extra[extraKey("provider")])
	}
}
