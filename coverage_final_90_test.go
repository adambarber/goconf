package goconf

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatOfUnknownCoverage(t *testing.T) {
	if _, ok := formatOf(Extension(".nope")); ok {
		t.Fatal("expected unknown")
	}
}

func TestPresenceOfUnknownFormatCoverage(t *testing.T) {
	_, err := presenceOf(SourceBytes("x"), Format("nope"), FilePath("f.nope"), reflect.TypeFor[intExtrasDoc]())
	if err == nil {
		t.Fatal("expected")
	}
}

func TestDecodeFormatUnknownCoverage(t *testing.T) {
	var v intExtrasDoc

	err := decodeFormat(SourceBytes("x"), FilePath("f.nope"), Target{Value: reflect.ValueOf(&v).Elem()}, Format("nope"))
	if err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLWantKindCoverage(t *testing.T) {
	_ = yamlWantKind(reflect.TypeFor[int](), "k")
	_ = yamlWantKind(reflect.TypeFor[bool](), "k")
	_ = yamlWantKind(reflect.TypeFor[[]int](), "k")
	_ = yamlWantKind(reflect.TypeFor[map[string]int](), "k")
	_ = yamlWantKind(reflect.TypeFor[struct{}](), "k")
	_ = yamlWantKind(reflect.TypeFor[string](), "k")
}

func TestYAMLUnknownFieldOfCoverage(t *testing.T) {
	te := &yaml.TypeError{Errors: []string{"line 1: field foo not found in type goconf.bar"}}
	_ = yamlUnknownFieldOf(te)
	te = &yaml.TypeError{Errors: []string{"other error"}}
	_ = yamlUnknownFieldOf(te)
	te = &yaml.TypeError{Errors: nil}
	_ = yamlUnknownFieldOf(te)
}

func TestExtrasFieldCoverage(t *testing.T) {
	if _, ok := extrasField(reflect.TypeFor[intExtrasDoc]()); !ok {
		t.Fatal("expected extras")
	}

	type noExtras struct {
		A resName `json:"a"`
	}

	if _, ok := extrasField(reflect.TypeFor[noExtras]()); ok {
		t.Fatal("unexpected")
	}
}

func TestCopyMissingCoverage(t *testing.T) {
	type cfg struct {
		A resName `json:"a"`
		B resName `json:"b"`
	}

	dst := reflect.ValueOf(&cfg{A: "x", B: ""}).Elem()
	def := reflect.ValueOf(cfg{A: "y", B: "z"})

	if err := copyMissing(dst, def, "", KeySet{paths: nil}); err != nil {
		t.Fatal(err)
	}

	if dst.FieldByName("B").String() != "z" {
		t.Fatalf("B=%v", dst.FieldByName("B").Interface())
	}
}
