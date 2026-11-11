package goconf

import (
	"path/filepath"
	"testing"
)

type mstr string

func TestLoadWithMetaJSONBasic(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\n  \"region\": \"eu\",\n  \"nested\": {\n    \"host\": \"h\",\n    \"port\": 1\n  }\n}\n")},
	})
	t.Chdir(string(root))

	type nested struct {
		Host mstr `json:"host"`
		Port int  `json:"port"`
	}

	type cfg struct {
		Region mstr   `json:"region"`
		Nested nested `json:"nested"`
	}

	got, prov, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertJSONBasicValues(t, got.Region, got.Nested.Host, got.Nested.Port, prov, meta)

	_ = filepath.Base
}

func assertJSONBasicValues(t *testing.T, region, host mstr, port int, prov []Provenance, meta Meta) {
	t.Helper()

	if region != "eu" || host != "h" || port != 1 {
		t.Fatalf("got region=%q host=%q port=%d", region, host, port)
	}

	if len(prov) != 1 || meta.Len() == 0 {
		t.Fatalf("prov=%d meta=%d", len(prov), meta.Len())
	}

	es := meta.BySchemaPath("region")
	if len(es) != 1 || es[0].Line != 2 || es[0].File != prov[0].Path || es[0].SourceIndex != 0 {
		t.Fatalf("region entry=%#v prov=%#v", es, prov)
	}
}
