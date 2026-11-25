package goconf

import (
	"os"
	"path/filepath"
	"testing"
)

// formatCrossedCfg declares different names per format. In every format both
// names are valid and distinct fields; presence must keep them distinct.
type formatCrossedCfg struct {
	Port       int `json:"port"        toml:"port"        yaml:"listen_port"`
	ListenPort int `json:"listen_port" toml:"listen_port" yaml:"port"`
}

// loadFormatCrossed writes one source in its own format and asserts both
// explicit values beat defaults that would otherwise fill them.
func loadFormatCrossed[T comparable](t *testing.T, f Format, name FileName, data SourceBytes, def T) T {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, string(name)), data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, _, meta, err := LoadWithMeta[T](Spec{Name: configName("app"), Format: f},
		WithSearchPaths(DirPath(dir)), WithDefaults(def))
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	for _, path := range []SchemaPath{"port", "listen_port"} {
		if len(meta.BySchemaPath(path)) != 1 {
			t.Fatalf("schema path %q must hold exactly one field: %#v", path, meta.Entries())
		}
	}

	return got
}

// Cross-format tags must not collapse two distinct JSON fields into one.
func TestFormatScopedPresenceJSONKeepsDistinctFields(t *testing.T) {
	got := loadFormatCrossed(t, FormatJSON, fileNameJSON,
		SourceBytes(`{"port":1,"listen_port":2}`), formatCrossedCfg{Port: 9, ListenPort: 9})
	if got.Port != 1 || got.ListenPort != 2 {
		t.Fatalf("explicit JSON values overwritten: Port=%d ListenPort=%d, want 1 and 2", got.Port, got.ListenPort)
	}
}

// Cross-format tags must not collapse two distinct YAML fields into one.
func TestFormatScopedPresenceYAMLKeepsDistinctFields(t *testing.T) {
	got := loadFormatCrossed(t, FormatYAML, fileNameYAML,
		SourceBytes("listen_port: 1\nport: 2\n"), formatCrossedCfg{Port: 9, ListenPort: 9})
	if got.Port != 1 || got.ListenPort != 2 {
		t.Fatalf("explicit YAML values overwritten: Port=%d ListenPort=%d, want 1 and 2", got.Port, got.ListenPort)
	}
}

// Cross-format tags must not regress TOML, whose own path walker was correct.
func TestFormatScopedPresenceTOMLKeepsDistinctFields(t *testing.T) {
	got := loadFormatCrossed(t, FormatTOML, fileNameTOML,
		SourceBytes("port = 1\nlisten_port = 2\n"), formatCrossedCfg{Port: 9, ListenPort: 9})
	if got.Port != 1 || got.ListenPort != 2 {
		t.Fatalf("explicit TOML values overwritten: Port=%d ListenPort=%d, want 1 and 2", got.Port, got.ListenPort)
	}
}
