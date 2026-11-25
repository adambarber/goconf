package goconf

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLPopulateMapNodeRootScalarMismatch(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("x\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), ""); err == nil {
		t.Fatal("expected")
	}
}
