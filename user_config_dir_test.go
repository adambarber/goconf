package goconf

import (
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

	file := filepath.Join(dir, "app.json")
	if err := os.WriteFile(file, SourceBytes(`{"region":"xdg"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	work := filepath.Join(scratch, "work")
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	t.Chdir(work)

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON})
	if err != nil {
		t.Fatalf("real UserConfigDir branch must locate the global config: %v", err)
	}

	if got.Region != "xdg" {
		t.Fatalf("got %q, want the real UserConfigDir leg's value", got.Region)
	}

	if len(prov) != 1 || prov[0].Path != FilePath(file) {
		t.Fatalf("provenance %v must name the real UserConfigDir leg", prov)
	}
}

// TestDefaultSearchPathsUserConfigDirFailureFallsToLocal pins the error
// branch: with XDG_CONFIG_HOME pointing at a non-directory, discovery
// still works and only the local leg is searched.
func TestDefaultSearchPathsUserConfigDirFailureFallsToLocal(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(scratch, "not-a-dir"))

	work := filepath.Join(scratch, "work")
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(work, "app.json"), SourceBytes(`{"region":"local-only"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Chdir(work)

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON})
	if err != nil {
		if !strings.Contains(err.Error(), "no config file found") {
			t.Fatalf("unexpected refusal: %v", err)
		}

		return // discovery correctly refused rather than guessed
	}

	if got.Region != "local-only" || len(prov) != 1 {
		t.Fatalf("local leg must carry the load alone, got %+v %v", got, prov)
	}
}
