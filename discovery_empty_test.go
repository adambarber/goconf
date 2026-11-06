package goconf

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestDiscoverExtensionBoundary: a file whose name merely extends the
// config name is not a candidate; the next character must be a dot.
func TestDiscoverExtensionBoundary(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {FileName("apple.json"): SourceBytes(`{"region":"sibling"}`)},
	})
	paths := []DirPath{DirPath(filepath.Join(string(root), string(dirLocal)))}

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: undetermined})
	if err == nil {
		t.Fatalf("apple.json must never be a candidate for name app; discovery must refuse")
	}

	if !strings.Contains(err.Error(), string(paths[0])) {
		t.Errorf("refusal %q must name the searched path, not adopt the sibling", err)
	}
}

// TestDiscoverEmptyRefusal: a load that found nothing loaded nothing.
// Zero files after discovery is an error naming every searched path.
func TestDiscoverEmptyRefusal(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	paths := []DirPath{
		DirPath(filepath.Join(string(root), string(dirGlobal))),
		DirPath(filepath.Join(string(root), string(dirLocal))),
	}

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("empty discovery must be refused")
	}

	for _, p := range paths {
		if !strings.Contains(err.Error(), string(p)) {
			t.Errorf("refusal %q must name searched path %q", err, p)
		}
	}
}
