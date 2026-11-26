package goconf

import (
	"reflect"
	"testing"
)

func TestFormatOfKnownCoverage(t *testing.T) {
	if _, ok := formatOf(Extension(".json")); !ok {
		t.Fatal("expected json")
	}

	if _, ok := formatOf(Extension(".yaml")); !ok {
		t.Fatal("expected yaml")
	}

	if _, ok := formatOf(Extension(".toml")); !ok {
		t.Fatal("expected toml")
	}
}

func TestDecodeFormatAndPresenceOfJSONYAMLTOML(t *testing.T) {
	type doc struct {
		A resName `json:"a" toml:"a" yaml:"a"`
	}

	var v doc
	if err := decodeFormat(SourceBytes(`{"a":"x"}`), FilePath("f.json"), Target{Value: reflect.ValueOf(&v).Elem()}, FormatJSON); err != nil {
		t.Fatal(err)
	}

	v = doc{A: ""}
	if err := decodeFormat(SourceBytes("a = \"x\"\n"), FilePath("f.toml"), Target{Value: reflect.ValueOf(&v).Elem()}, FormatTOML); err != nil {
		t.Fatal(err)
	}

	v = doc{A: ""}
	if err := decodeFormat(SourceBytes("a: x\n"), FilePath("f.yaml"), Target{Value: reflect.ValueOf(&v).Elem()}, FormatYAML); err != nil {
		t.Fatal(err)
	}
}

func TestYAMLWantKindPointerCoverage(t *testing.T) {
	_ = yamlWantKind(reflect.TypeFor[*int](), "k")
	_ = yamlWantKind(reflect.TypeFor[**string](), "k")
}

func TestLineAtByteBounds(t *testing.T) {
	if n := lineAtByte(SourceBytes(""), 5); n != 1 {
		t.Fatalf("n=%d", n)
	}

	if n := lineAtByte(SourceBytes("ab"), 10); n < 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestConsumeJSONObjectEndOK(t *testing.T) {
	assertJSONObjectEnd(t, SourceBytes(`{"a":1}`), false)
}

func TestYAMLPopulateMapNodeNestedMapping(t *testing.T) {
	got := loadBytesOK[yamlMapDoc](t, SourceBytes("m:\n  a: b\n"), FormatYAML)
	if got.M["a"] != "b" {
		t.Fatalf("got=%v", got)
	}
}
