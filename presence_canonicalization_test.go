package goconf

import "testing"

// portUntagged has no tags, so the schema spelling is the Go field name.
type portUntagged struct {
	Port int
}

// portTagged declares different names per format; the schema spelling is the
// first declared tag, json's "port".
type portTagged struct {
	Port int `json:"port" toml:"port" yaml:"listen_port"`
}

// portCaseTagged declares a lowercase json name; the document may uppercase it.
type portCaseTagged struct {
	Port int `json:"port" toml:"port" yaml:"port"`
}

// nestedPortCfg and nestedPort form the ordinary nested-struct control case.
type nestedPortCfg struct {
	Server nestedPort `json:"server" toml:"server" yaml:"server"`
}

type nestedPort struct {
	Port int `json:"port" toml:"port" yaml:"port"`
}

// portFileFor maps one format to the file discovery expects for that format.
func portFileFor(f Format) FileName {
	switch f {
	case FormatJSON:
		return fileNameJSON
	case FormatYAML:
		return fileNameYAML
	case FormatTOML:
		return fileNameTOML
	case undetermined:
		return FileName("app.unknown")
	default:
		return FileName("app.unknown")
	}
}

// loadWithPortDefault loads one real file through the public API with a
// default that would win if the explicit key were mistaken for absence.
func loadWithPortDefault[T comparable](t *testing.T, f Format, data SourceBytes, def T) (T, Meta) {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {portFileFor(f): data},
	})

	got, _, meta, err := LoadWithMeta[T](Spec{Name: configName("app"), Format: f},
		append(loadOpts(globalPathOf(root)), WithDefaults(def))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	return got, meta
}

// TestExplicitUntaggedJSONKeyBeatsDefault pins the first confirmed variant:
// an untagged Go field named Port, written as lowercase "port" in JSON.
func TestExplicitUntaggedJSONKeyBeatsDefault(t *testing.T) {
	got, _ := loadWithPortDefault(t, FormatJSON, SourceBytes(`{"port":9090}`), portUntagged{Port: 8080})
	if got.Port != 9090 {
		t.Fatalf("explicit JSON value lost to default: got %d, want 9090", got.Port)
	}
}

// TestExplicitUntaggedYAMLKeyBeatsDefault pins the untagged mismatch in YAML.
func TestExplicitUntaggedYAMLKeyBeatsDefault(t *testing.T) {
	got, _ := loadWithPortDefault(t, FormatYAML, SourceBytes("port: 9090\n"), portUntagged{Port: 8080})
	if got.Port != 9090 {
		t.Fatalf("explicit YAML value lost to default: got %d, want 9090", got.Port)
	}
}

// TestExplicitYAMLKeyBeatsDefaultWithDivergentTags pins the third variant:
// the document uses the yaml tag, while the schema vocabulary uses the json
// tag. The document spelling must survive only as metadata.
func TestExplicitYAMLKeyBeatsDefaultWithDivergentTags(t *testing.T) {
	got, meta := loadWithPortDefault(t, FormatYAML, SourceBytes("listen_port: 9090\n"), portTagged{Port: 8080})
	if got.Port != 9090 {
		t.Fatalf("explicit YAML value lost to default: got %d, want 9090", got.Port)
	}

	bySchema := meta.BySchemaPath("port")
	if len(bySchema) != 1 || bySchema[0].Key != "listen_port" {
		t.Fatalf("schema path must be canonical and key must keep document spelling: %#v", meta.Entries())
	}
}

// TestExplicitCaseInsensitiveJSONKeyBeatsDefault pins the fourth variant:
// the strict decoder accepts uppercase, so presence must accept it too.
func TestExplicitCaseInsensitiveJSONKeyBeatsDefault(t *testing.T) {
	got, meta := loadWithPortDefault(t, FormatJSON, SourceBytes(`{"PORT":9090}`), portCaseTagged{Port: 8080})
	if got.Port != 9090 {
		t.Fatalf("explicit case-insensitive JSON value lost to default: got %d, want 9090", got.Port)
	}

	byPath := meta.ByPath("port")
	if len(byPath) != 1 || byPath[0].Key != "PORT" {
		t.Fatalf("canonical path with document-spelling key: %#v", meta.Entries())
	}
}

// TestExplicitTOMLKeyBeatsDefault is the positive control: TOML already
// canonicalized presence, so it must keep behaving the same after the fix.
func TestExplicitTOMLKeyBeatsDefault(t *testing.T) {
	got, _ := loadWithPortDefault(t, FormatTOML, SourceBytes("port = 9090\n"), portUntagged{Port: 8080})
	if got.Port != 9090 {
		t.Fatalf("explicit TOML value lost to default: got %d, want 9090", got.Port)
	}
}

// TestExplicitNestedJSONKeyBeatsDefault pins the ordinary nested-struct case:
// canonicalization must descend through a struct without losing leaf
// presence or indexed metadata.
func TestExplicitNestedJSONKeyBeatsDefault(t *testing.T) {
	def := nestedPortCfg{Server: nestedPort{Port: 8080}}

	got, meta := loadWithPortDefault(t, FormatJSON, SourceBytes(`{"server":{"port":9090}}`), def)
	if got.Server.Port != 9090 {
		t.Fatalf("explicit nested value lost to default: got %d, want 9090", got.Server.Port)
	}

	if len(meta.BySchemaPath("server.port")) != 1 || len(meta.ByPath("server.port")) != 1 {
		t.Fatalf("nested presence missing: %#v", meta.Entries())
	}
}

// extrasCfg pins the "-" exclusion: the extras map is not a schema field,
// so an arbitrary key keeps its document spelling in metadata.
type extrasCfg struct {
	Extra map[mstr]mstr `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  mstr          `json:"name"     toml:"name" yaml:"name"`
}

func TestExtrasKeyKeepsDocumentSpellingInMeta(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"name":"n","extra":"x"}`)},
	})

	got, _, meta, err := LoadWithMeta[extrasCfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if got.Extra["extra"] != "x" {
		t.Fatalf("extras value missing: %#v", got.Extra)
	}

	bySchema := meta.BySchemaPath("extra")
	if len(bySchema) != 1 || bySchema[0].Key != "extra" {
		t.Fatalf("extras spelling must survive: %#v", meta.Entries())
	}
}

func TestUnknownJSONKeyStillRefusedWithCanonicalizedPresence(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"nope":1}`)},
	})

	_, _, err := Load[portUntagged](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	mustRefuseNames(t, err, "unknown field")
}
