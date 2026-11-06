package goconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDiscoverRefusesStatFailure: a searchable-but-unsearchable
// directory must surface its stat failure as a refusal naming the
// candidate path, never as silent absence. Skipped as root.
func TestDiscoverRefusesStatFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not bind root")
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
	})
	dir := DirPath(filepath.Join(string(root), string(dirLocal)))

	const restoreBits os.FileMode = 0o700 // user rwx: TempDir cleanup must traverse and unlink

	if err := os.Chmod(string(dir), 0o400); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(string(dir), restoreBits); err != nil {
			t.Logf("cleanup chmod: %v", err)
		}
	})

	_, err := discover(Discovery{Search: []DirPath{dir}, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("stat failure must be refused, not treated as absence")
	}

	for _, need := range []testText{testText(fileNameJSON), testText("permission denied")} {
		if !strings.Contains(err.Error(), string(need)) {
			t.Errorf("refusal %q must name %q", err, need)
		}
	}
}
