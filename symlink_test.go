package goconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadRefusesSymlinkEscape: a candidate inside a searched directory
// whose target leaves the searched directory is a refusal naming the file
// and where it points; confinement, not a documented non-goal.
func TestLoadRefusesSymlinkEscape(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {},
	})

	secret := filepath.Join(string(root), "secret.json")
	if err := os.WriteFile(secret, SourceBytes(`{"region":"SECRET"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	link := filepath.Join(string(root), string(dirLocal), string(fileNameJSON))
	if err := os.Symlink(secret, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(filepath.Join(string(root), string(dirLocal)))))
	if err == nil {
		t.Fatalf("symlink escape must be refused")
	}

	if !strings.Contains(err.Error(), string(fileNameJSON)) {
		t.Errorf("refusal %q must name the file", err)
	}

	if !strings.Contains(err.Error(), secret) {
		t.Errorf("refusal %q must name where the link points", err)
	}
}

// TestLoadReadsInTreeSymlink: a symlink whose target stays inside the
// searched directory is a legitimate candidate and still loads.
func TestLoadReadsInTreeSymlink(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {},
	})

	data := filepath.Join(string(root), string(dirLocal), "data.json")
	if err := os.WriteFile(data, SourceBytes(`{"region":"eu"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	link := filepath.Join(string(root), string(dirLocal), string(fileNameJSON))
	if err := os.Symlink(data, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(filepath.Join(string(root), string(dirLocal)))))
	if err != nil {
		t.Fatalf("in-dir symlink must load: %v", err)
	}

	if got.Region != "eu" {
		t.Fatalf("got %q, want eu", got.Region)
	}
}
