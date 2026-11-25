package goconf

import (
	"reflect"
	"testing"
)

func TestYAMLPopulateMapNodeRootScalarMismatch(t *testing.T) {
	root := yamlContentNode(t, SourceText("x\n"))

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), ""); err == nil {
		t.Fatal("expected")
	}
}
