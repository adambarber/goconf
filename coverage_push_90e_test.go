package goconf

import (
	"reflect"
	"testing"
)

func TestYAMLPopulateMapNodeRootMismatch(t *testing.T) {
	root := yamlContentNode(t, SourceText("- a\n"))

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), ""); err == nil {
		t.Fatal("expected")
	}
}

func TestYAMLWantKindMoreCoverage(t *testing.T) {
	_ = yamlWantKind(reflect.TypeFor[uint](), "k")
	_ = yamlWantKind(reflect.TypeFor[float64](), "k")
	_ = yamlWantKind(reflect.TypeFor[[2]int](), "k")
}
