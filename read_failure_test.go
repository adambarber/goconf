package goconf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadReadFailureNamesPathAndLeg: an unreadable candidate must be
// refused naming its full path and the search leg it came from, never a
// bare read error. Skipped as root, which bypasses permission bits.
func TestLoadReadFailureNamesPathAndLeg(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not bind root")
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
		dirLocal:  {fileNameJSON: SourceBytes(`{"region":"us"}`)},
	})
	local := writeFileAt(t, DirPath(filepath.Join(string(root), string(dirLocal))), fileNameJSON, SourceBytes(`{"region":"us"}`))
	stripAccess(t, local, 0o000, 0o600)

	chdirLocal(t, root)

	err := loadRefused(t, Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(globalPathOf(root), legPaths(root, dirLocal)[0]))

	mustRefuseNames(t, err, string(local), "local")
}
