package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestAbsenceWithoutOptionStillRefused pins the unchanged behavior: the
// option is opt-in, absence without it keeps the ErrNothingFound refusal.
func TestAbsenceWithoutOptionStillRefused(t *testing.T) {
	paths := optionalAbsencePaths(t)

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(paths...))
	if err == nil {
		t.Fatalf("absence without the option must be refused")
	}

	if !errors.Is(err, ErrNothingFound) {
		t.Fatalf("refusal %v must errors.Is ErrNothingFound", err)
	}
}

// TestUnlistableDirWithOptionStillRefused pins the fault boundary: an
// unlistable search directory is a fault, never absence, even with the
// option set.
func TestUnlistableDirWithOptionStillRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	t.Chdir(string(root))

	unlistable := filepath.Join(string(root), "unlistable")
	if err := os.MkdirAll(unlistable, 0); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.Chmod(unlistable, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	defer func() {
		if rmErr := os.Remove(unlistable); rmErr != nil {
			t.Logf("cleanup remove: %v", rmErr)
		}
	}()

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(unlistable)), WithOptionalConfig())
	if err == nil {
		t.Fatalf("an unlistable directory is a fault, not absence; must refuse")
	}

	assertReasonIO(t, err, caseName("unlistable with option"))
}

// TestInvalidConfigWithOptionStillRefused pins the invalid boundary:
// present-but-invalid config refuses even with the option set.
func TestInvalidConfigWithOptionStillRefused(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameJSON: SourceBytes(`{"region":"eu","extra":1}`),
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...), WithOptionalConfig())
	if err == nil {
		t.Fatalf("present-but-invalid config must refuse even with the option")
	}

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("refusal %v must errors.Is ErrUnknownField", err)
	}
}

// TestPresentFileWithOptionNoWarning pins the quiet case: a file was
// found, so the option changes nothing and prints no warning.
func TestPresentFileWithOptionNoWarning(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameJSON: SourceBytes(`{"region":"eu"}`),
	})
	t.Chdir(string(root))
	buf := captureWarn(t)

	got, prov := loadOK(t, Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...), WithOptionalConfig())

	cfgEquals(t, got, loadCfgWant(0, "eu", nestedCfgWant("", 0)))

	if len(prov) != 1 {
		t.Fatalf("got %d provenance entries, want 1", len(prov))
	}

	if got := buf.String(); got != "" {
		t.Fatalf("a found file must print no warning, got %q", got)
	}
}

// TestOptionalWarningGolden pins the warning text byte-exact, including
// the trailing newline the standard log package appends.
func TestOptionalWarningGolden(t *testing.T) {
	paths := optionalAbsencePaths(t)
	buf := captureWarn(t)

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(paths...), WithOptionalConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	golden := "[warn] No config file found. Using default values.\n"
	if buf.String() != golden {
		t.Fatalf("warning bytes differ:\n got %q\nwant %q", buf.String(), golden)
	}
}
