package goconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test directory names shared across the discovery tests.
const (
	dirGlobal = searchDir("global")
	dirLocal  = searchDir("local")
)

// searchDir is one directory searched for config files.
type searchDir string

// Test file names shared across the discovery tests.
const (
	fileNameJSON = FileName("app.json")
	fileNameYAML = FileName("app.yaml")
	fileNameTOML = FileName("app.toml")
)

// writeTree creates files under a fresh temp root and returns the root.
func writeTree(t *testing.T, files map[searchDir]map[FileName]SourceBytes) DirPath {
	t.Helper()

	root := t.TempDir()

	for dir, entries := range files {
		full := filepath.Join(root, string(dir))
		if err := os.MkdirAll(full, 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		for name, data := range entries {
			if err := os.WriteFile(filepath.Join(full, string(name)), data, 0o600); err != nil {
				t.Fatalf("write: %v", err)
			}
		}
	}

	return DirPath(root)
}

func TestDiscoverGlobalThenLocal(t *testing.T) {
	data := SourceBytes(`{"region":"eu"}`)
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: data},
		dirLocal: {fileNameJSON: SourceBytes(`{"region":"us"}`)},
	})
	paths := []DirPath{DirPath(filepath.Join(string(root), "global")), DirPath(filepath.Join(string(root), "local"))}

	found, err := discover(Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(found), found)
	}

	if found[0].Path != FilePath(filepath.Join(string(root), "global", string(fileNameJSON))) {
		t.Errorf("global file must come first: %v", found[0].Path)
	}
}

func TestDiscoverFormatNarrowsSearch(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu"}`), fileNameYAML: SourceBytes("region: eu\n")},
	})
	paths := []DirPath{DirPath(filepath.Join(string(root), "global"))}

	found, err := discover(Discovery{Search: paths, Name: configName("app"), Format: FormatYAML})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 1 || found[0].Path != FilePath(filepath.Join(string(root), "global", string(fileNameYAML))) {
		t.Fatalf("got %v, want only app.yaml", found)
	}
}

func TestDiscoverRefusesAmbiguousDirectory(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{}`), fileNameTOML: SourceBytes("")},
	})
	paths := []DirPath{DirPath(filepath.Join(string(root), "global"))}

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: undetermined})
	if err == nil {
		t.Fatalf("expected ambiguous refusal, got nil")
	}

	for _, need := range []testText{testText(root), testText(fileNameJSON), testText(fileNameTOML)} {
		if !strings.Contains(err.Error(), string(need)) {
			t.Errorf("refusal %q must name %q", err, need)
		}
	}
}

func TestDiscoverRefusesInventedFormat(t *testing.T) {
	_, err := discover(Discovery{Search: []DirPath{"."}, Name: configName("app"), Format: Format(".yml")})

	if err == nil || !strings.Contains(err.Error(), ".yml") || !strings.Contains(err.Error(), acceptedFormats) {
		t.Fatalf("expected refusal naming .yml and accepted spellings, got %v", err)
	}
}

// testText is text inside a discovery test assertion.
type testText string

