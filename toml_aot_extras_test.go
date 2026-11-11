package goconf

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type (
	tomlAOTTestName   string
	tomlAOTTestExpect string
)

const (
	tomlAOTExtrasNameA      tomlAOTTestName   = "a"
	tomlAOTExtrasNameB      tomlAOTTestName   = "b"
	tomlAOTExtrasProvider   extraKey          = "provider"
	tomlAOTExtrasRegion     extraKey          = "region"
	tomlAOTExtrasProviderA  tomlAOTTestExpect = "aws"
	tomlAOTExtrasProviderB  tomlAOTTestExpect = "gcp"
	tomlAOTExtrasRegionA    tomlAOTTestExpect = "us-east-1"
	tomlAOTExtrasPathItems  OccurrencePath    = "items"
	tomlAOTExtrasSchemaItem SchemaPath        = "items"
)

type tomlAOTExtrasItem struct {
	Extra map[extraKey]extraVal `goconf:"extras" toml:"-"`
	Name  tomlAOTTestName       `toml:"name"`
}

type tomlAOTExtrasDoc struct {
	Items []tomlAOTExtrasItem `toml:"items"`
}

func writeTOMLAOTExtrasConfig(t *testing.T) (DirPath, FilePath) {
	t.Helper()

	dir := t.TempDir()
	src := "[[items]]\nname = \"a\"\nprovider = \"aws\"\nregion = \"us-east-1\"\n\n[[items]]\nname = \"b\"\nprovider = \"gcp\"\n"
	file := filepath.Join(dir, "app.toml")

	if err := os.WriteFile(file, []byte(src), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return DirPath(dir), FilePath(file)
}

func TestLoadTOMLArrayOfTablesExtrasCapturedPerElement(t *testing.T) {
	dir, _ := writeTOMLAOTExtrasConfig(t)

	got, prov, err := Load[tomlAOTExtrasDoc](Spec{Name: Name("app"), Format: FormatTOML}, WithSearchPaths(dir))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(prov) != 1 || len(got.Items) != 2 {
		t.Fatalf("unexpected result: prov=%d items=%d", len(prov), len(got.Items))
	}

	assertTOMLAOTExtrasDecoded(t, got)
}

func assertTOMLAOTExtrasDecoded(t *testing.T, got tomlAOTExtrasDoc) {
	t.Helper()

	if got.Items[0].Name != tomlAOTExtrasNameA || got.Items[1].Name != tomlAOTExtrasNameB {
		t.Fatalf("names = %q,%q", got.Items[0].Name, got.Items[1].Name)
	}

	assertTOMLAOTExtraValue(t, got.Items[0].Extra, tomlAOTExtrasProvider, extraVal(tomlAOTExtrasProviderA))
	assertTOMLAOTExtraValue(t, got.Items[0].Extra, tomlAOTExtrasRegion, extraVal(tomlAOTExtrasRegionA))
	assertTOMLAOTExtraValue(t, got.Items[1].Extra, tomlAOTExtrasProvider, extraVal(tomlAOTExtrasProviderB))

	if len(got.Items[0].Extra) != 2 || len(got.Items[1].Extra) != 1 {
		t.Fatalf("extras lens = %d,%d (%#v %#v)", len(got.Items[0].Extra), len(got.Items[1].Extra), got.Items[0].Extra, got.Items[1].Extra)
	}
}

func assertTOMLAOTExtraValue(t *testing.T, extra map[extraKey]extraVal, key extraKey, want extraVal) {
	t.Helper()

	if extra[key] != want {
		t.Fatalf("extra[%q] = %#v, want %#v (all=%#v)", key, extra[key], want, extra)
	}
}

func TestLoadWithMetaTOMLArrayOfTablesExtrasPerElement(t *testing.T) {
	dir, file := writeTOMLAOTExtrasConfig(t)

	got, prov, meta, err := LoadWithMeta[tomlAOTExtrasDoc](Spec{Name: Name("app"), Format: FormatTOML}, WithSearchPaths(dir))
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(prov) != 1 || len(got.Items) != 2 {
		t.Fatalf("unexpected result: prov=%d items=%d", len(prov), len(got.Items))
	}

	assertTOMLAOTExtrasDecoded(t, got)
	assertTOMLAOTExtrasMeta(t, meta, file)
}

func assertTOMLAOTExtrasMeta(t *testing.T, meta Meta, file FilePath) {
	t.Helper()

	want := tomlAOTExtrasWantEntries()
	entries := meta.Entries()

	if len(entries) != len(want) {
		t.Fatalf("meta entries = %d, want %d (%#v)", len(entries), len(want), entries)
	}

	for i, w := range want {
		assertTOMLAOTMetaEntry(t, i, &entries[i], w, file)
	}

	assertTOMLAOTExtrasMetaValues(t, meta)
}

type tomlAOTMetaWant struct {
	path   OccurrencePath
	schema SchemaPath
	key    OccurrenceKey
	line   int
	eff    bool
}

func tomlAOTExtrasWantEntries() []tomlAOTMetaWant {
	return []tomlAOTMetaWant{
		{path: tomlAOTExtrasPathItems, schema: tomlAOTExtrasSchemaItem, key: "items", line: 1, eff: true},
		{path: "items[0]", schema: tomlAOTExtrasSchemaItem, key: "[0]", line: 1, eff: true},
		{path: "items[0].name", schema: "items.name", key: "name", line: 2, eff: true},
		{path: "items[0].provider", schema: "items.provider", key: OccurrenceKey(tomlAOTExtrasProvider), line: 3, eff: false},
		{path: "items[0].region", schema: "items.region", key: OccurrenceKey(tomlAOTExtrasRegion), line: 4, eff: false},
		{path: "items[1]", schema: tomlAOTExtrasSchemaItem, key: "[1]", line: 6, eff: true},
		{path: "items[1].name", schema: "items.name", key: "name", line: 7, eff: true},
		{path: "items[1].provider", schema: "items.provider", key: OccurrenceKey(tomlAOTExtrasProvider), line: 8, eff: false},
	}
}

func assertTOMLAOTMetaEntry(t *testing.T, i int, got *Occurrence, want tomlAOTMetaWant, file FilePath) {
	t.Helper()

	if got.Path != want.path || got.SchemaPath != want.schema || got.Key != want.key || got.Line != want.line || got.Effective != want.eff {
		t.Fatalf("entry[%d] = %#v, want path=%q schema=%q key=%q line=%d eff=%v", i, got, want.path, want.schema, want.key, want.line, want.eff)
	}

	if got.File != file || got.SourceIndex != 0 {
		t.Fatalf("entry[%d] file/source = (%q,%d), want (%q,0)", i, got.File, got.SourceIndex, file)
	}
}

func assertTOMLAOTExtrasMetaValues(t *testing.T, meta Meta) {
	t.Helper()

	assertMetaPathValue(t, meta, "items[0].name", tomlAOTTestExpect(tomlAOTExtrasNameA))
	assertMetaPathValue(t, meta, "items[0].provider", tomlAOTExtrasProviderA)
	assertMetaPathValue(t, meta, "items[0].region", tomlAOTExtrasRegionA)
	assertMetaPathValue(t, meta, "items[1].provider", tomlAOTExtrasProviderB)

	if len(meta.BySchemaPath("items.provider")) != 2 {
		t.Fatalf("schema items.provider entries = %#v, want 2", meta.BySchemaPath("items.provider"))
	}
}

func assertMetaPathValue(t *testing.T, meta Meta, path OccurrencePath, want tomlAOTTestExpect) {
	t.Helper()

	entries := meta.ByPath(path)
	if len(entries) != 1 || !entries[0].Value.IsValid() {
		t.Fatalf("%s value = %#v, want %q", path, entries, want)
	}

	got := entries[0].Value
	if got.Kind() == reflect.String {
		if got.String() != string(want) {
			t.Fatalf("%s value = %q, want %q", path, got.String(), want)
		}

		return
	}

	if got.Interface() != string(want) {
		t.Fatalf("%s value = %#v, want %q", path, got.Interface(), want)
	}
}
