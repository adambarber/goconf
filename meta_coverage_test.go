package goconf

import "testing"

func TestLoadWithMetaYAMLArrayOccurrences(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("items:\n  - name: a\n  - name: b\n")},
	})
	t.Chdir(string(root))

	type item struct {
		Name mstr `yaml:"name"`
	}

	type cfg struct {
		Items []item `yaml:"items"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	const pathItems0Name OccurrencePath = "items[0].name"

	const pathItems1Name OccurrencePath = "items[1].name"

	sn := meta.BySchemaPath("items.name")
	if len(sn) != 2 || sn[0].Path != pathItems0Name || sn[1].Path != pathItems1Name {
		t.Fatalf("entries=%#v", meta.Entries())
	}

	if sn[0].Line != 2 || sn[1].Line != 3 {
		t.Fatalf("lines=%#v", sn)
	}
}

func TestLoadWithMetaExtrasTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("name = \"n\"\nextra = \"x\"\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Extra map[mstr]mstr `goconf:"extras" toml:"-"`
		Name  mstr          `toml:"name"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(meta.BySchemaPath("extra")) != 1 {
		t.Fatalf("entries=%#v", meta.Entries())
	}
}

func TestLoadWithMetaDynamicYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("env:\n  A: \"1\"\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Env map[mstr]mstr `goconf:"dynamic" yaml:"env"`
	}

	_, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(meta.ByPath("env.A")) != 1 {
		t.Fatalf("entries=%#v", meta.Entries())
	}
}

func TestLoadWithMetaDefaultsEffective(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\"region\":\"eu\"}\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Region  mstr `json:"region"`
		Retries int  `json:"retries"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatJSON},
		append(loadOpts(globalPathOf(root)), WithDefaults(cfg{Region: "d", Retries: 5}))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if got.Region != "eu" || got.Retries != 5 {
		t.Fatalf("got=%+v", got)
	}

	rs := meta.BySchemaPath("region")
	if len(rs) != 1 || !rs[0].Effective || !rs[0].Value.IsValid() || rs[0].Value.String() != "eu" {
		t.Fatalf("region=%#v", rs)
	}
}
