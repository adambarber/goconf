package goconf

import (
	"path/filepath"
	"testing"
)

// TestLoadRefusesSymlinkEscape: a candidate inside a searched directory
// whose target leaves the searched directory is a refusal naming the file
// and where it points; confinement, not a documented non-goal.
func TestLoadRefusesSymlinkEscape(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{})

	secret := writeFileAt(t, root, "secret.json", SourceBytes(`{"region":"SECRET"}`))
	symlinkCandidate(t, root, fileNameJSON, secret)

	err := loadRefused(t, Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...))

	mustRefuseNames(t, err, string(fileNameJSON), string(secret))
}

// TestLoadReadsInTreeSymlink: a symlink whose target stays inside the
// searched directory is a legitimate candidate and still loads.
func TestLoadReadsInTreeSymlink(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{})

	data := writeFileAt(t, DirPath(filepath.Join(string(root), string(dirLocal))), "data.json", SourceBytes(`{"region":"eu"}`))
	symlinkCandidate(t, root, fileNameJSON, data)

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("in-dir symlink must load: %v", err)
	}

	if got.Region != "eu" {
		t.Fatalf("got %q, want eu", got.Region)
	}
}
