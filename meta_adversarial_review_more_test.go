package goconf

import "testing"

func TestLoadWithMetaYAMLMergeSequence(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("b1: &b1\n  x: 1\n  y: 1\nb2: &b2\n  y: 2\n  z: 2\na:\n  <<: [*b1, *b2]\n  z: 3\n")},
	})
	t.Chdir(string(root))

	type inner struct {
		X int `yaml:"x"`
		Y int `yaml:"y"`
		Z int `yaml:"z"`
	}

	type cfg struct {
		B1 inner `yaml:"b1"`
		B2 inner `yaml:"b2"`
		A  inner `yaml:"a"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if got.A.X != 1 || got.A.Y != 1 || got.A.Z != 3 {
		t.Fatalf("a=%#v", got.A)
	}

	mustMetaPath(t, meta, "a.x", "a.x")
	mustMetaPath(t, meta, "a.y", "a.y")
	mustMetaPath(t, meta, "a.z", "a.z")
}

func TestLoadWithMetaTOMLLineTricky(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("# name = \"comment\"\nname = \"real\"\nother = \"name = fake\"\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Name  mstr `toml:"name"`
		Other mstr `toml:"other"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	name := meta.BySchemaPath("name")
	if len(name) != 1 || name[0].Line != 2 {
		t.Fatalf("name entry=%#v all=%#v", name, meta.Entries())
	}

	other := meta.BySchemaPath("other")
	if len(other) != 1 || other[0].Line != 3 {
		t.Fatalf("other entry=%#v all=%#v", other, meta.Entries())
	}
}

type mrinner struct {
	A int `json:"a" toml:"a" yaml:"a"`
	B int `json:"b" toml:"b" yaml:"b"`
}

type mrinline struct {
	Items []mrinner `json:"items" toml:"items" yaml:"items"`
}

func TestLoadWithMetaTOMLInlineTableArrayElements(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("items = [{a=1,b=2},{a=3,b=4}]\n")},
	})
	t.Chdir(string(root))

	got, _, meta, err := LoadWithMeta[mrinline](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(got.Items) != 2 || got.Items[0].A != 1 || got.Items[0].B != 2 || got.Items[1].A != 3 || got.Items[1].B != 4 {
		t.Fatalf("items=%#v", got.Items)
	}

	mustMetaPath(t, meta, "items[0].a", "items.a")
	mustMetaPath(t, meta, "items[0].b", "items.b")
	mustMetaPath(t, meta, "items[1].a", "items.a")
	mustMetaPath(t, meta, "items[1].b", "items.b")
}

func mustMetaPath(t *testing.T, meta Meta, path OccurrencePath, schema SchemaPath) {
	t.Helper()

	es := meta.ByPath(path)
	if len(es) == 0 {
		t.Fatalf("missing path %q entries=%#v", path, meta.Entries())
	}

	if es[0].SchemaPath != schema || !es[0].Effective {
		t.Fatalf("path %q got schema=%q eff=%v want schema=%q eff=true entry=%#v", path, es[0].SchemaPath, es[0].Effective, schema, es[0])
	}
}
