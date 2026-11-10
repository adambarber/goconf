package goconf

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func TestTOMLCaptureKeyEmpty(t *testing.T) {
	ok, err := tomlCaptureKey(reflect.Value{}, reflect.Value{}, toml.Key{}, FilePath("f.toml"))
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestYAMLPopulateStructNodeMismatchNonRoot(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("- a\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	dst := reflect.New(reflect.TypeFor[intExtrasDoc]()).Elem()
	if err := yamlPopulateStructNode(root, dst, FilePath("f.yaml"), "nested"); err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLPopulateMapNodeAliasAndDefault(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("a: 1\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

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
	dec := json.NewDecoder(strings.NewReader(`{"a":1`))
	dec.UseNumber()

	if _, err := dec.Token(); err != nil {
		t.Fatal(err)
	}

	if _, err := decodeJSONObjectKey(dec, SourceBytes(`{"a":1`), FilePath("f.json")); err != nil {
		t.Fatal(err)
	}

	tok, err := dec.Token()
	if err != nil {
		t.Fatal(err)
	}

	_ = tok

	if err := consumeJSONObjectEnd(dec, SourceBytes(`{"a":1`), FilePath("f.json")); err == nil {
		t.Fatal("expected")
	}
}
