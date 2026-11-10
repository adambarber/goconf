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

func TestCopyMissingWithPresentField(t *testing.T) {
	type cfg struct {
		A resName `json:"a"`
		B resName `json:"b"`
	}

	dst := reflect.ValueOf(&cfg{A: "x", B: ""}).Elem()
	def := reflect.ValueOf(cfg{A: "y", B: "z"})
	present := KeySet{paths: nil}.Add("a")

	if err := copyMissing(dst, def, "", present); err != nil {
		t.Fatal(err)
	}

	if dst.FieldByName("A").String() != "x" {
		t.Fatalf("A=%v", dst.FieldByName("A").Interface())
	}
}
