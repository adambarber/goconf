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
	got := loadBytesOK[intExtrasDoc](t, SourceBytes(`{"name":"r","n":42}`), FormatJSON)
	if got.Extra["n"] != 42 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedBoolJSON(t *testing.T) {
	got := loadBytesOK[boolExtrasDoc](t, SourceBytes(`{"name":"r","flag":true}`), FormatJSON)
	if !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedFloatJSON(t *testing.T) {
	got := loadBytesOK[floatExtrasDoc](t, SourceBytes(`{"name":"r","f":1.5}`), FormatJSON)
	if got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedUintJSON(t *testing.T) {
	got := loadBytesOK[uintExtrasDoc](t, SourceBytes(`{"name":"r","u":7}`), FormatJSON)
	if got.Extra["u"] != 7 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedSliceJSON(t *testing.T) {
	got := loadBytesOK[sliceExtrasDoc](t, SourceBytes(`{"name":"r","items":["a","b"]}`), FormatJSON)
	if len(got.Extra["items"]) != 2 || got.Extra["items"][0] != "a" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedIntYAML(t *testing.T) {
	got := loadBytesOK[intExtrasDoc](t, SourceBytes("name: r\nn: 42\n"), FormatYAML)
	if got.Extra["n"] != 42 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedBoolYAML(t *testing.T) {
	got := loadBytesOK[boolExtrasDoc](t, SourceBytes("name: r\nflag: true\n"), FormatYAML)
	if !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedFloatYAML(t *testing.T) {
	got := loadBytesOK[floatExtrasDoc](t, SourceBytes("name: r\nf: 1.5\n"), FormatYAML)
	if got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedSliceYAML(t *testing.T) {
	got := loadBytesOK[sliceExtrasDoc](t, SourceBytes("name: r\nitems:\n  - a\n  - b\n"), FormatYAML)
	if len(got.Extra["items"]) != 2 || got.Extra["items"][0] != "a" {
		t.Fatalf("got=%v", got)
	}
}
