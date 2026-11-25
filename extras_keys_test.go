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
	got := loadResourceOK(t, SourceBytes(`{"resource":{"name":"r","kind":"k","provider":"`+testProvider+`"}}`), FormatJSON)

	assertResourceExtras(t, got)
}

func TestExtrasYAML(t *testing.T) {
	got := loadResourceOK(t, SourceBytes("resource:\n  name: r\n  kind: k\n  provider: aws\n"), FormatYAML)

	assertResourceExtras(t, got)
}

func TestExtrasTOML(t *testing.T) {
	got := loadResourceOK(t, SourceBytes("[resource]\nname = \"r\"\nkind = \"k\"\nprovider = \"aws\"\n"), FormatTOML)

	assertResourceExtras(t, got)
}

func TestExtrasFixedFieldPrecedenceJSON(t *testing.T) {
	got := loadResourceOK(t, SourceBytes(`{"resource":{"name":"fixed","kind":"k","provider":"aws"}}`), FormatJSON)

	if got.Resource.Name != "fixed" {
		t.Fatalf("name = %q", got.Resource.Name)
	}

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}

func TestUnknownOutsideExtrasStillRefused(t *testing.T) {
	refuseBytes[loadCfg](t, SourceBytes(`{"region":"eu","b":"y"}`), FormatJSON, WantText("expected refusal"))
}

func TestExtrasYAMLPrecedence(t *testing.T) {
	got := loadResourceOK(t, SourceBytes("resource:\n  name: fixed\n  kind: k\n  provider: aws\n"), FormatYAML)

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}

func TestExtrasTOMLPrecedence(t *testing.T) {
	got := loadResourceOK(t, SourceBytes("[resource]\nname = \"fixed\"\nkind = \"k\"\nprovider = \"aws\"\n"), FormatTOML)

	if _, ok := got.Resource.Extra["name"]; ok {
		t.Fatalf("fixed field must not be captured as extra")
	}
}
