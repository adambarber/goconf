package goconf

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestJSONExtrasRootShapeMoreCoverage(t *testing.T) {
	if err := jsonExtrasRootShape(FilePath("f.json"), '0', reflect.ValueOf(struct{}{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), 't', reflect.ValueOf([]int{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), 'n', reflect.ValueOf(struct{}{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), '[', reflect.ValueOf(map[string]int{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), '{', reflect.ValueOf([1]int{})); err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLPopulateStructNodeTypeMismatchNested(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("1\n"), &doc); err != nil {
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

func TestYAMLPopulateMapNodeTypeMismatchNested(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("1\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), "nested"); err == nil {
		t.Fatal("expected")
	}
}

func TestTOMLSetExtraValueConvertibleInt(t *testing.T) {
	extras := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraInt8]()))
	key := reflect.ValueOf(extraKey("k"))
	val := reflect.ValueOf(int64(7))

	if err := setExtraValue(extras, key, val, FilePath("f.toml"), "k"); err != nil {
		t.Fatal(err)
	}
}
