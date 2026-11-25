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
	got := loadBytesOK[dynExtrasDoc](t, SourceBytes("[items.one]\nname = \"n\"\nx = \"y\"\n"), FormatTOML)

	if got.Items["one"].Name != "n" || got.Items["one"].Extra["x"] != "y" {
		t.Fatalf("got=%#v", got)
	}
}

func TestYAMLExtrasInsideDynamicMapEntry(t *testing.T) {
	got := loadBytesOK[dynExtrasDoc](t, SourceBytes("items:\n  one:\n    name: n\n    x: y\n"), FormatYAML)

	if got.Items["one"].Extra["x"] != "y" {
		t.Fatalf("got=%#v", got)
	}
}

func TestJSONExtrasInsideDynamicMapEntry(t *testing.T) {
	got := loadBytesOK[dynExtrasDoc](t, SourceBytes(`{"items":{"one":{"name":"n","x":"y"}}}`), FormatJSON)

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
	err := refuseBytesErr[jsonUnknownExtrasDoc](t, SourceBytes(`{"in":{"name":"n","bad":1}}`), FormatJSON)
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("err=%v", err)
	}
}

func TestYAMLRootScalarExtrasNotMapping(t *testing.T) {
	err := refuseBytesErr[intExtrasDoc](t, SourceBytes("just-a-scalar\n"), FormatYAML)
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrNotAMapping) {
		t.Fatalf("err=%v", err)
	}
}
