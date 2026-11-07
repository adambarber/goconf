package goconf

import (
	"os"
	"testing"
)

// TestDiscoverRefusesStatFailure: a searchable-but-unsearchable
// directory must surface its stat failure as a refusal naming the
// candidate path, never as silent absence. Skipped as root.
func TestDiscoverRefusesStatFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not bind root")
	}

	root := oneLegTree(t, map[FileName]SourceBytes{fileNameJSON: SourceBytes(`{"region":"eu"}`)})
	dir := legPaths(root, dirLocal)[0]

	// user rwx: TempDir cleanup must traverse and unlink
	stripAccess(t, FilePath(dir), 0o400, 0o700)

	_, err := discover(Discovery{Search: []DirPath{dir}, Name: configName("app"), Format: FormatJSON})

	mustRefuseNames(t, err, string(fileNameJSON), "permission denied")
}
