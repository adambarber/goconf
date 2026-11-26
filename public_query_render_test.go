package goconf

import (
	"strings"
	"testing"
)

type queryRegion string

type queryCfg struct {
	Region queryRegion `json:"region" toml:"region" yaml:"region"`
}

func loadLayeredQueryMeta(t *testing.T) Meta {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"global"}`)},
		dirLocal:  {fileNameJSON: SourceBytes(`{"region":"local"}`)},
	})

	_, _, meta, err := LoadWithMeta[queryCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirGlobal, dirLocal)...))
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	return meta
}

func assertLookupAliases(t *testing.T, meta Meta) {
	t.Helper()

	path := OccurrencePath("region")
	if got, want := meta.Lookup(path), meta.ByPath(path); len(got) != len(want) || len(got) == 0 {
		t.Fatalf("Lookup = %d entries, ByPath = %d", len(got), len(want))
	}

	schema := SchemaPath("region")
	if got, want := meta.LookupSchema(schema), meta.BySchemaPath(schema); len(got) != len(want) || len(got) == 0 {
		t.Fatalf("LookupSchema = %d entries, BySchemaPath = %d", len(got), len(want))
	}
}

func assertIndexedAliases(t *testing.T, meta Meta) {
	t.Helper()

	at, ok := meta.At(OccurrencePath("region"), 1)
	if !ok || at.Path != "region" {
		t.Fatalf("At = %+v, ok = %v", at, ok)
	}

	atSchema, ok := meta.AtSchema(SchemaPath("region"), 1)
	if !ok || atSchema.SchemaPath != "region" {
		t.Fatalf("AtSchema = %+v, ok = %v", atSchema, ok)
	}

	effective := meta.Effective()
	if len(effective) == 0 || len(effective) > len(meta.Entries()) {
		t.Fatalf("Effective = %d entries, entries = %d", len(effective), len(meta.Entries()))
	}

	for _, entry := range effective {
		if !entry.Effective {
			t.Fatalf("Effective returned a non-effective entry: %+v", entry)
		}
	}
}

// TestMetaQueryAliasesAgree pins the convenience query methods against the
// canonical ones through a real layered load.
func TestMetaQueryAliasesAgree(t *testing.T) {
	meta := loadLayeredQueryMeta(t)

	assertLookupAliases(t, meta)
	assertIndexedAliases(t, meta)
}

type renderKey string

type mapRenderCfg struct {
	Items map[renderKey]secretText `json:"items" secret:"true"`
	Plain map[renderKey]mstr       `json:"plain"`
}

const redactProbeSecret = "hunter2"

// TestRedactRendersMapFields pins the map rendering path: keys stay visible,
// marked values are redacted, and ordering is stable.
func TestRedactRendersMapFields(t *testing.T) {
	cfg := mapRenderCfg{
		Items: map[renderKey]secretText{"b": redactProbeSecret, "a": "letmein"},
		Plain: map[renderKey]mstr{"z": "last", "y": "first"},
	}

	rendered := string(Redact(cfg))

	if !strings.Contains(rendered, "map[") {
		t.Fatalf("Redact = %q, want a rendered map", rendered)
	}

	if strings.Contains(rendered, redactProbeSecret) || strings.Contains(rendered, "letmein") {
		t.Fatalf("Redact echoed a marked map value: %q", rendered)
	}

	if !strings.Contains(rendered, "[redacted]") {
		t.Fatalf("Redact = %q, want redacted map values", rendered)
	}

	if strings.Index(rendered, "y:first") > strings.Index(rendered, "z:last") {
		t.Fatalf("Redact map keys are not sorted: %q", rendered)
	}
}
