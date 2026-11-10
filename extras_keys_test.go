package goconf

import "testing"

type extraKey string

// extraVal holds a captured extras value as text for test assertions.
type extraVal string

type (
	resName string
	resKind string
)

type resource struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName               `json:"name"     toml:"name" yaml:"name"`
	Kind  resKind               `json:"kind"     toml:"kind" yaml:"kind"`
}

type resourceDoc struct {
	Resource resource `json:"resource" toml:"resource" yaml:"resource"`
}

const testProvider = "aws"

func TestExtrasJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"resource":{"name":"r","kind":"k","provider":"` + testProvider + `"}}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Resource.Name != "r" || got.Resource.Kind != "k" {
		t.Fatalf("fixed: %#v", got.Resource)
	}

	if got.Resource.Extra[extraKey("provider")] != extraVal(testProvider) {
		t.Fatalf("extra provider = %#v", got.Resource.Extra[extraKey("provider")])
	}
}

func TestExtrasYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("resource:\n  name: r\n  kind: k\n  provider: aws\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Resource.Name != "r" || got.Resource.Kind != "k" {
		t.Fatalf("fixed: %#v", got.Resource)
	}

	if got.Resource.Extra[extraKey("provider")] != extraVal(testProvider) {
		t.Fatalf("extra provider = %#v", got.Resource.Extra[extraKey("provider")])
	}
}

func TestExtrasTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("[resource]\nname = \"r\"\nkind = \"k\"\nprovider = \"aws\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Resource.Name != "r" || got.Resource.Kind != "k" {
		t.Fatalf("fixed: %#v", got.Resource)
	}

	if got.Resource.Extra[extraKey("provider")] != extraVal(testProvider) {
		t.Fatalf("extra provider = %#v", got.Resource.Extra[extraKey("provider")])
	}
}

func TestExtrasFixedFieldPrecedenceJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"resource":{"name":"fixed","kind":"k","provider":"aws"}}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Resource.Name != "fixed" {
		t.Fatalf("name = %q", got.Resource.Name)
	}

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}

type strictA string

func TestUnknownOutsideExtrasStillRefused(t *testing.T) {
	type strict struct {
		A strictA `json:"a" toml:"a" yaml:"a"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"a":"x","b":"y"}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[strict](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestExtrasYAMLPrecedence(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("resource:\n  name: fixed\n  kind: k\n  provider: aws\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}

func TestExtrasTOMLPrecedence(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("[resource]\nname = \"fixed\"\nkind = \"k\"\nprovider = \"aws\"\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[resourceDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}
