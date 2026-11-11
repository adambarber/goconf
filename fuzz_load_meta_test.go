package goconf

import (
	"strings"
	"testing"
)

func FuzzLoadWithMeta(f *testing.F) {
	seeds := []string{
		`{"region":"eu","nested":{"host":"h","port":1}}`,
		"region: eu\nnested:\n  host: h\n  port: 1\n",
		"region = \"eu\"\n\n[nested]\nhost = \"h\"\nport = 1\n",
		`{"items":[{"name":"a"},{"name":"b"}]}`,
		"items:\n  - name: a\n  - name: b\n",
		"[[items]]\nname = \"a\"\n[[items]]\nname = \"b\"\n",
		`{"a":{"b":1}}`,
		"a = [1,2,3]\n",
		"base: &b\n  x: 1\na:\n  <<: *b\n  y: 2\n",
		``,
		"\x00\x01",
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, format := range []Format{FormatJSON, FormatYAML, FormatTOML} {
			runFuzzMetaLeg(t, data, format)
		}
	})
}

func runFuzzMetaLeg(t *testing.T, data []byte, format Format) {
	t.Helper()

	dir := t.TempDir()
	_ = writeFileAt(t, DirPath(dir), FileName("app"+string(format)), SourceBytes(data))

	_, prov, meta, err := LoadWithMeta[fuzzMetaCfg](Spec{Name: configName("app"), Format: format},
		WithSearchPaths(DirPath(dir)))
	if err != nil {
		return
	}

	entries := meta.Entries()
	if len(entries) != meta.Len() {
		t.Fatalf("entries len %d != meta len %d", len(entries), meta.Len())
	}

	assertFuzzMetaEntries(t, entries, len(prov), meta)
}

func assertFuzzMetaEntries(t *testing.T, entries []Occurrence, provLen int, meta Meta) {
	t.Helper()

	for i := range entries {
		assertFuzzMetaEntry(t, i, &entries[i], provLen)
	}

	for _, e := range entries {
		if len(meta.ByPath(e.Path)) == 0 {
			t.Fatalf("ByPath missing %q", e.Path)
		}

		if len(meta.BySchemaPath(e.SchemaPath)) == 0 {
			t.Fatalf("BySchemaPath missing %q", e.SchemaPath)
		}
	}
}

func assertFuzzMetaEntry(t *testing.T, i int, e *Occurrence, provLen int) {
	t.Helper()

	if e.SourceIndex < 0 || e.SourceIndex >= provLen {
		t.Fatalf("entry %d bad source index %d prov=%d", i, e.SourceIndex, provLen)
	}

	if e.Line < 1 {
		t.Fatalf("entry %d bad line %d path=%q", i, e.Line, e.Path)
	}

	if e.File == "" {
		t.Fatalf("entry %d empty file path=%q", i, e.Path)
	}

	if e.Path == "" {
		t.Fatalf("entry %d empty path", i)
	}

	if e.SchemaPath == "" {
		t.Fatalf("entry %d empty schema path path=%q", i, e.Path)
	}

	assertFuzzMetaKeyPath(t, i, e)
}

func assertFuzzMetaKeyPath(t *testing.T, i int, e *Occurrence) {
	t.Helper()

	p := string(e.Path)
	if len(p) == 0 || p[len(p)-1] != ']' {
		return
	}

	j := strings.LastIndexByte(p, '[')
	if j < 0 || j == len(p)-1 {
		return
	}

	suffix := p[j:]
	if suffix != string(e.Key) {
		t.Fatalf("entry %d key/path mismatch path=%q key=%q", i, e.Path, e.Key)
	}
}

type fuzzMetaCfg struct {
	Extra  map[mstr]mstr    `goconf:"extras" json:"-"      toml:"-"      yaml:"-"`
	Region mstr             `goconf:"-"      json:"region" toml:"region" yaml:"region"`
	Nested fuzzMetaNested   `goconf:"-"      json:"nested" toml:"nested" yaml:"nested"`
	Items  []fuzzMetaNested `goconf:"-"      json:"items"  toml:"items"  yaml:"items"`
	A      fuzzMetaNested   `goconf:"-"      json:"a"      toml:"a"      yaml:"a"`
	Base   fuzzMetaNested   `goconf:"-"      json:"base"   toml:"base"   yaml:"base"`
}

type fuzzMetaNested struct {
	Host mstr `goconf:"-" json:"host" toml:"host" yaml:"host"`
	Name mstr `goconf:"-" json:"name" toml:"name" yaml:"name"`
	Port int  `goconf:"-" json:"port" toml:"port" yaml:"port"`
	X    int  `goconf:"-" json:"x"    toml:"x"    yaml:"x"`
	Y    int  `goconf:"-" json:"y"    toml:"y"    yaml:"y"`
	B    int  `goconf:"-" json:"b"    toml:"b"    yaml:"b"`
}
