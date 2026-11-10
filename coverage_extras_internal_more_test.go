package goconf

import (
	"encoding/json/jsontext"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLExtrasRootNotMappingInternal(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("42\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestTOMLRawValueTermCoverage(t *testing.T) {
	cases := []reflect.Value{
		{},
		reflect.ValueOf(map[string]int{"a": 1}),
		reflect.ValueOf([]int{1}),
		reflect.ValueOf(true),
		reflect.ValueOf("s"),
		reflect.ValueOf(int64(1)),
		reflect.ValueOf(uint64(1)),
		reflect.ValueOf(float64(1.0)),
	}
	for _, c := range cases {
		_ = tomlRawValueTerm(c)
	}
}

func TestJSONRootTokenKindCoverage(t *testing.T) {
	_, _ = jsonRootKindName('[')
	_, _ = jsonRootKindName('{')
	_, _ = jsonRootKindName('t')
	_, _ = jsonRootKindName('0')
	_, _ = jsonRootKindName('"')
	_, _ = jsonRootKindName('n')
}

func TestSetJSONBoolMismatchCoverage(t *testing.T) {
	dst := reflect.New(reflect.TypeFor[int64]()).Elem()
	if err := setJSONBool(dst, true, FilePath("f.json"), "x"); err == nil {
		t.Fatal("expected")
	}
}

func TestSetJSONScalarMismatchCoverage(t *testing.T) {
	dst := reflect.New(reflect.TypeFor[int64]()).Elem()
	if err := setJSONScalar(dst, jsonText("hi"), FilePath("f.json"), "x"); err == nil {
		t.Fatal("expected")
	}
}

func TestDecodeJSONObjectUnknownFieldRefusalCoverage(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`{"a":1}`), jsontext.AllowDuplicateNames(true))

	if _, err := dec.ReadToken(); err != nil {
		t.Fatal(err)
	}

	key, err := decodeJSONObjectKey(dec, SourceBytes(`{"a":1}`), FilePath("f.json"))
	if err != nil {
		t.Fatal(err)
	}

	ref := jsonUnknownFieldRefusal(SourceBytes(`{"a":1}`), FilePath("f.json"), dec, key)
	if ref == nil {
		t.Fatal("expected refusal")
	}
}

func TestConsumeJSONObjectEndCoverage(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(`{}`), jsontext.AllowDuplicateNames(true))

	if _, err := dec.ReadToken(); err != nil {
		t.Fatal(err)
	}

	if err := consumeJSONObjectEnd(dec, SourceBytes(`{}`), FilePath("f.json")); err != nil {
		t.Fatal(err)
	}
}

func TestYAMLPopulateSliceNodeMismatch(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("a: 1\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	dst := reflect.New(reflect.TypeFor[[]extraVal]()).Elem()
	if err := yamlPopulateSliceNode(root, dst, FilePath("f.yaml"), "x"); err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLPopulateSliceCoverage(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("- a\n- b\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	dst := reflect.New(reflect.TypeFor[[]extraVal]()).Elem()
	if err := yamlPopulateSlice(root, dst, FilePath("f.yaml"), "x"); err != nil {
		t.Fatal(err)
	}

	if dst.Len() != 2 {
		t.Fatalf("len=%d", dst.Len())
	}
}
