package goconf

import "testing"

func TestLoadWithMetaYAMLBasic(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("region: eu\nnested:\n  host: h\n  port: 1\n")},
	})
	t.Chdir(string(root))

	type nested struct {
		Host mstr `json:"host" yaml:"host"`
		Port int  `json:"port" yaml:"port"`
	}

	type cfg struct {
		Region mstr   `json:"region" yaml:"region"`
		Nested nested `json:"nested" yaml:"nested"`
	}

	_, prov, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	es := meta.BySchemaPath("nested.host")
	if len(es) != 1 {
		t.Fatalf("nested.host entries=%d all=%v", len(es), meta.Entries())
	}

	if es[0].Line != 3 || es[0].File != prov[0].Path {
		t.Fatalf("entry=%+v", es[0])
	}
}

func TestLoadWithMetaJSONArrayOccurrences(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\n  \"items\": [\n    {\"name\": \"a\"},\n    {\"name\": \"b\"}\n  ]\n}\n")},
	})
	t.Chdir(string(root))

	type item struct {
		Name mstr `json:"name"`
	}

	type cfg struct {
		Items []item `json:"items"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(got.Items) != 2 || got.Items[0].Name != "a" || got.Items[1].Name != "b" {
		t.Fatalf("got %+v", got)
	}

	assertJSONArrayOccurrences(t, meta)
}

func assertJSONArrayOccurrences(t *testing.T, meta Meta) {
	t.Helper()

	if len(meta.ByPath("items[0]")) != 1 || len(meta.ByPath("items[1]")) != 1 {
		t.Fatalf("element paths: %#v", meta.Entries())
	}

	const pathItems0Name OccurrencePath = "items[0].name"

	const pathItems1Name OccurrencePath = "items[1].name"

	sn := meta.BySchemaPath("items.name")
	if len(sn) != 2 || sn[0].Path != pathItems0Name || sn[1].Path != pathItems1Name {
		t.Fatalf("schema order/paths: %#v", sn)
	}

	if occ, ok := meta.AtSchema("items.name", 1); !ok || occ.Path != pathItems1Name {
		t.Fatalf("AtSchema: %+v %v", occ, ok)
	}
}

func TestLoadWithMetaMergeLayers(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\"region\":\"eu\"}\n")},
		dirLocal:  {fileNameJSON: SourceBytes("{\"region\":\"us\"}\n")},
	})
	chdirLocal(t, root)

	type cfg struct {
		Region mstr `json:"region"`
	}

	got, prov, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(legPaths(root, dirGlobal)[0])...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertMergeLayersMeta(t, got.Region, prov, meta)
}

func assertMergeLayersMeta(t *testing.T, region mstr, prov []Provenance, meta Meta) {
	t.Helper()

	if region != "us" || len(prov) != 2 {
		t.Fatalf("region=%q prov=%d", region, len(prov))
	}

	rs := meta.BySchemaPath("region")
	if len(rs) != 2 || rs[0].SourceIndex != 0 || rs[1].SourceIndex != 1 {
		t.Fatalf("region occs=%#v", rs)
	}

	if rs[0].Effective || !rs[1].Effective || !rs[1].Value.IsValid() || rs[1].Value.String() != "us" {
		t.Fatalf("effective: %#v", rs)
	}
}

func TestLoadWithMetaDynamicAndExtras(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\"env\":{\"A\":\"1\"},\"extra\":\"x\",\"name\":\"n\"}\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Env   map[mstr]mstr `goconf:"dynamic" json:"env"`
		Extra map[mstr]mstr `goconf:"extras"  json:"-"`
		Name  mstr          `json:"name"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(meta.ByPath("env.A")) != 1 {
		t.Fatalf("dynamic key missing: %#v", meta.Entries())
	}

	if len(meta.ByPath("extra")) == 0 && len(meta.BySchemaPath("extra")) == 0 {
		t.Fatalf("extras key missing: %#v", meta.Entries())
	}
}

func TestLoadWithMetaTOMLAOT(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("[[items]]\nname = \"a\"\n\n[[items]]\nname = \"b\"\n")},
	})
	t.Chdir(string(root))

	type item struct {
		Name mstr `toml:"name"`
	}

	type cfg struct {
		Items []item `toml:"items"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(got.Items) != 2 || got.Items[0].Name != "a" || got.Items[1].Name != "b" {
		t.Fatalf("got %+v", got)
	}

	assertTOMLAOTMeta(t, meta)
}

func assertTOMLAOTMeta(t *testing.T, meta Meta) {
	t.Helper()

	sn := meta.BySchemaPath("items.name")
	if len(sn) != 2 || sn[0].Path != "items[0].name" || sn[1].Path != "items[1].name" {
		t.Fatalf("paths=%#v all=%#v", sn, meta.Entries())
	}

	if sn[0].Line <= 0 || sn[1].Line <= 0 || sn[1].Line <= sn[0].Line {
		t.Fatalf("lines: %#v %#v", sn[0], sn[1])
	}
}
