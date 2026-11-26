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
	got := loadBytesOK[ptrExtrasDoc](t, SourceBytes(`{"inner":{"name":"n"},"x":"y"}`), FormatJSON)
	if got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasPointerStructYAML(t *testing.T) {
	got := loadBytesOK[ptrExtrasDoc](t, SourceBytes("inner:\n  name: n\nx: y\n"), FormatYAML)
	if got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasPointerStructTOML(t *testing.T) {
	got := loadBytesOK[ptrExtrasDoc](t, SourceBytes("x = \"y\"\n[inner]\nname = \"n\"\n"), FormatTOML)
	if got.Inner == nil || got.Inner.Name != "n" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v", got)
	}
}

type mapFieldDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-" toml:"-" yaml:"-"`
	M     map[extraKey]extraVal `json:"m"        toml:"m" yaml:"m"`
}

func TestExtrasMapFieldYAML(t *testing.T) {
	got := loadBytesOK[mapFieldDoc](t, SourceBytes("m:\n  a: b\nx: y\n"), FormatYAML)
	if got.M["a"] != "b" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasMapFieldJSON(t *testing.T) {
	got := loadBytesOK[mapFieldDoc](t, SourceBytes(`{"m":{"a":"b"},"x":"y"}`), FormatJSON)
	if got.M["a"] != "b" || got.Extra["x"] != "y" {
		t.Fatalf("got=%v", got)
	}
}
