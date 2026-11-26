package goconf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRefusesAbsoluteSymlinkTarget pins os.Root's relative-target rule:
// confinement cannot be bypassed with an absolute link into the same tree.
func TestLoadRefusesAbsoluteSymlinkTarget(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{})
	local := DirPath(filepath.Join(string(root), string(dirLocal)))
	data := writeFileAt(t, local, "data.json", SourceBytes(`{"port":7}`))

	link := filepath.Join(string(local), string(fileNameJSON))
	if err := os.Symlink(string(data), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	err := loadRefused(t, Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...))

	mustRefuseNames(t, err, string(fileNameJSON), string(data))
}
