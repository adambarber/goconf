package goconf

import (
	"os"
	"path/filepath"
	"testing"
)

// symlinkCandidate links name in the local leg at target with a relative
// target, which is the only spelling os.Root accepts.
func symlinkCandidate(t *testing.T, root DirPath, name FileName, target FilePath) {
	t.Helper()

	link := filepath.Join(string(root), string(dirLocal), string(name))

	rel, err := filepath.Rel(filepath.Dir(link), string(target))
	if err != nil {
		t.Fatalf("relative symlink: %v", err)
	}

	if err := os.Symlink(rel, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
}
