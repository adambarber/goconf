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

type mapExtrasDoc struct {
	Extra map[extraKey]mapVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName             `json:"name"     toml:"name" yaml:"name"`
}

type mapVal map[extraKey]extraVal

func TestExtrasStructValueJSON(t *testing.T) {
	got := loadBytesOK[structExtrasDoc](t, SourceBytes(`{"name":"r","s":{"x":"v"}}`), FormatJSON)
	if got.Extra["s"].X != "v" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasStructValueYAML(t *testing.T) {
	got := loadBytesOK[structExtrasDoc](t, SourceBytes("name: r\ns:\n  x: v\n"), FormatYAML)
	if got.Extra["s"].X != "v" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasStructValueTOMLRefused(t *testing.T) {
	refuseBytes[structExtrasDoc](t, SourceBytes("name = \"r\"\n[s]\nx = \"v\"\n"), FormatTOML, WantText("expected refusal"))
}

func TestExtrasFloat32TOML(t *testing.T) {
	got := loadBytesOK[float32ExtrasDoc](t, SourceBytes("name = \"r\"\nf = 1.5\n"), FormatTOML)
	if got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasJSONArrayInTypedSliceJSON(t *testing.T) {
	got := loadBytesOK[sliceExtrasDoc](t, SourceBytes(`{"name":"r","items":["a","b","c"]}`), FormatJSON)
	if len(got.Extra["items"]) != 3 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasYAMLMapValue(t *testing.T) {
	got := loadBytesOK[mapExtrasDoc](t, SourceBytes("name: r\nm:\n  a: b\n"), FormatYAML)
	if got.Extra["m"]["a"] != "b" {
		t.Fatalf("got=%v", got)
	}
}
