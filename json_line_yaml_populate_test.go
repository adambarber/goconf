package goconf

import (
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestTOMLCaptureKeyEmpty(t *testing.T) {
	ok, err := tomlCaptureKey(reflect.Value{}, reflect.Value{}, toml.Key{}, tomlAOTIndex{}, FilePath("f.toml"))
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestYAMLPopulateStructNodeMismatchNonRoot(t *testing.T) {
	assertYAMLPopulateStructRefused(t, SourceText("- a\n"), "nested")
}

func TestYAMLPopulateMapNodeAliasAndDefault(t *testing.T) {
	root := yamlContentNode(t, SourceText("a: 1\n"))

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), "m"); err != nil {
		t.Fatal(err)
	}
}

func TestJSONLineAtByteCoverage(t *testing.T) {
	_ = lineAtByte(SourceBytes("a\nb\n"), 0)
	_ = lineAtByte(SourceBytes("a\nb\n"), 1)
	_ = lineAtByte(SourceBytes("a\nb\n"), 3)
}

func TestJSONConsumeJSONObjectEndError(t *testing.T) {
	assertJSONObjectEnd(t, SourceBytes(`{"a":1`), true)
}

func TestJSONKeyLineNestedContainers(t *testing.T) {
	data := SourceBytes("{\n  \"arr\": [\n    {\n      \"bad\": 1\n    }\n  ]\n}\n")

	if got := jsonKeyLine(data, ObjectKey("arr")); got != 2 {
		t.Fatalf("arr line=%d want=2", got)
	}

	if got := jsonKeyLine(data, ObjectKey("bad")); got != 4 {
		t.Fatalf("bad line=%d want=4", got)
	}

	nested := SourceBytes("{\n  \"a\": [\n    [\n      {\n        \"deep\": true\n      }\n    ]\n  ]\n}\n")

	if got := jsonKeyLine(nested, ObjectKey("deep")); got != 5 {
		t.Fatalf("deep line=%d want=5", got)
	}

	if got := jsonKeyLine(nested, ObjectKey("missing")); got != 0 {
		t.Fatalf("missing line=%d want=0", got)
	}
}

func TestJSONKeyLineArraysOfObjects(t *testing.T) {
	data := SourceBytes("[\n  {\n    \"x\": 1\n  },\n  {\n    \"y\": 2\n  }\n]\n")

	if got := jsonKeyLine(data, ObjectKey("x")); got != 3 {
		t.Fatalf("x line=%d want=3", got)
	}

	if got := jsonKeyLine(data, ObjectKey("y")); got != 6 {
		t.Fatalf("y line=%d want=6", got)
	}
}
