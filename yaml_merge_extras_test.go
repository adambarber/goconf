package goconf

import "testing"

type yamlMergeRegion string

type yamlMergeBase struct {
	Region yamlMergeRegion `yaml:"region"`
	Port   int             `yaml:"port"`
}

type yamlMergeExtrasCfg struct {
	Extra  map[extraKey]extraVal `goconf:"extras" json:"-" toml:"-"`
	Region yamlMergeRegion       `yaml:"region"`
	Base   yamlMergeBase         `yaml:"base"`
	Port   int                   `yaml:"port"`
}

func TestYAMLMergeKeysApplyWithExtras(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		"app.yaml": SourceBytes("base: &b\n  port: 8080\n  region: us\nregion: eu\n<<: *b\n"),
	})

	got, _, meta, err := LoadWithMeta[yamlMergeExtrasCfg](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Port != 8080 || got.Region != yamlMergeRegion("eu") {
		t.Fatalf("merged config = %+v, want Port=8080 Region=eu", got)
	}

	if _, ok := got.Extra[extraKey("<<")]; ok {
		t.Fatalf("merge key was captured as an extra: %v", got.Extra)
	}

	if len(meta.BySchemaPath(SchemaPath("port"))) == 0 {
		t.Fatal("merged port has no occurrence metadata")
	}
}
