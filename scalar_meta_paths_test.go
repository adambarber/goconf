package goconf

import "testing"

type yamlScalarCfg struct {
	Flag  bool    `yaml:"flag"`
	Count int     `yaml:"count"`
	Ratio float64 `yaml:"ratio"`
}

// TestYAMLUntaggedScalarMetadata drives implicit YAML scalar values through
// the metadata collector, not just the decoder.
func TestYAMLUntaggedScalarMetadata(t *testing.T) {
	src := SourceBytes("flag: true\ncount: 7\nratio: 1.5\n")

	cfg, _, meta, err := LoadBytesWithMeta[yamlScalarCfg](src, Spec{Name: configName("app"), Format: FormatYAML})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !cfg.Flag || cfg.Count != 7 || cfg.Ratio != 1.5 {
		t.Fatalf("cfg = %+v, want true/7/1.5", cfg)
	}

	entries := meta.Entries()
	if len(entries) != 3 {
		t.Fatalf("entries = %+v, want three scalar occurrences", entries)
	}

	for _, entry := range entries {
		if entry.LineConfidence != LineExact {
			t.Fatalf("entry %+v lost exact line confidence", entry)
		}
	}
}

type tomlMetaItem struct {
	Extra map[extraKey]extraVal `goconf:"extras" toml:"-"`
	Name  tomlAOTTestName       `toml:"name"`
}

type tomlMetaNested struct {
	Tags []tomlMetaItem `toml:"tags"`
}

type tomlMetaCfg struct {
	Items  []tomlMetaItem `toml:"items"`
	Nested tomlMetaNested `toml:"nested"`
}

// TestTOMLNestedArrayMetadata drives AOT, inline array, and extras paths in
// the TOML metadata collector.
func TestTOMLNestedArrayMetadata(t *testing.T) {
	src := SourceBytes("[[items]]\nname = \"a\"\nextra = \"e\"\n\n[nested]\ntags = [{name = \"b\"}]\n")

	cfg, _, meta, err := LoadBytesWithMeta[tomlMetaCfg](src, Spec{Name: configName("app"), Format: FormatTOML})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(cfg.Items) != 1 || cfg.Items[0].Name != "a" {
		t.Fatalf("items = %+v, want one a", cfg.Items)
	}

	if len(cfg.Nested.Tags) != 1 || cfg.Nested.Tags[0].Name != "b" {
		t.Fatalf("tags = %+v, want one b", cfg.Nested.Tags)
	}

	entries := meta.Entries()
	if len(entries) < 4 {
		t.Fatalf("entries = %+v, want nested occurrences", entries)
	}

	for _, entry := range entries {
		if entry.LineConfidence != LineHeuristic {
			t.Fatalf("entry %+v lost heuristic line confidence", entry)
		}
	}
}
