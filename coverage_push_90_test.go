package goconf

import (
	"reflect"
	"testing"
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
	root := yamlContentNode(t, SourceText("1\n"))

	dst := reflect.New(reflect.TypeFor[intExtrasDoc]()).Elem()
	if err := yamlPopulateStructNode(root, dst, FilePath("f.yaml"), "nested"); err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLPopulateMapNodeTypeMismatchNested(t *testing.T) {
	root := yamlContentNode(t, SourceText("1\n"))

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
