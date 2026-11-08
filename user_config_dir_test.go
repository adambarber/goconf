package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultSearchPathsRealUserConfigDir exercises the real
// os.UserConfigDir branch with HOME pinned to a scratch dir. Hermetic:
// never reads the real user config.
func TestDefaultSearchPathsRealUserConfigDir(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("HOME", scratch)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(scratch, "xdg"))

	global, cfgErr := os.UserConfigDir() // resolves inside the scratch dir only
	if cfgErr != nil {
		t.Fatalf("UserConfigDir must resolve with HOME pinned to scratch: %v", cfgErr)
	}

	if !strings.HasPrefix(global, scratch) {
		t.Fatalf("UserConfigDir %q escaped the scratch root %q", global, scratch)
	}

	dir := filepath.Join(global, "app")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	writeFileAt(t, DirPath(dir), "app.json", SourceBytes(`{"region":"xdg"}`))

	scratchWorkDir(t, DirPath(scratch))

	got, prov := loadOK(t, Spec{Name: configName("app"), Format: FormatJSON})

	if got.Region != "xdg" {
		t.Fatalf("got %q, want the real UserConfigDir leg's value", got.Region)
	}

	if len(prov) != 1 || prov[0].Path != FilePath(filepath.Join(dir, "app.json")) {
		t.Fatalf("provenance %v must name the real UserConfigDir leg", prov)
	}
}

// TestDefaultSearchPathsUserConfigDirFailureFallsToLocal pins the error
// branch: with XDG_CONFIG_HOME pointing at a non-directory, discovery
// still works and only the local leg is searched.
func TestDefaultSearchPathsUserConfigDirFailureFallsToLocal(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(scratch, "not-a-dir"))

	scratchWorkDir(t, DirPath(scratch))

	writeFileAt(t, DirPath(filepath.Join(scratch, "work")), "app.json", SourceBytes(`{"region":"local-only"}`))

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON})
	if err != nil {
		if !errors.Is(err, ErrNothingFound) {
			t.Fatalf("unexpected refusal: %v", err)
		}

		return // discovery correctly refused rather than guessed
	}

	if got.Region != "local-only" || len(prov) != 1 {
		t.Fatalf("local leg must carry the load alone, got %+v %v", got, prov)
	}
}
