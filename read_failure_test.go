package goconf

import (
	"os"
	"path/filepath"
	"strings"
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

	local := filepath.Join(string(root), string(dirLocal), string(fileNameJSON))
	if err := os.Chmod(local, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(local, 0o600); err != nil {
			t.Logf("cleanup chmod: %v", err)
		}
	})

	t.Chdir(filepath.Join(string(root), string(dirLocal)))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(globalPathOf(root), DirPath(filepath.Join(string(root), string(dirLocal)))))
	if err == nil {
		t.Fatalf("unreadable candidate must be refused")
	}

	for _, need := range []testText{testText(local), testText("local")} {
		if !strings.Contains(err.Error(), string(need)) {
			t.Errorf("refusal %q must name %q", err, need)
		}
	}
}
