package goconf

import (
	"bytes"
	"log"
	"os"
	"testing"
)

// optionalAbsencePaths builds two readable but empty search directories:
// every leg lists fine and zero candidates exist — true absence.
func optionalAbsencePaths(t *testing.T) []DirPath {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{dirGlobal: {}, dirLocal: {}})

	return legPaths(root, dirGlobal, dirLocal)
}

// captureWarn redirects the standard log's output into a fresh buffer
// for the test's duration and restores stderr and the flags after.
func captureWarn(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(0)

	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	})

	return buf
}

// TestOptionalAbsenceLoadsDefaults pins the forgivable case: with the
// option, true absence loads — defaults fill, provenance is empty, and
// the warning line is printed exactly once.
func TestOptionalAbsenceLoadsDefaults(t *testing.T) {
	paths := optionalAbsencePaths(t)
	buf := captureWarn(t)

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(paths...), WithOptionalConfig(),
		WithDefaults(loadCfgWant(5, "fill", nestedCfgWant("dh", 9))))
	if err != nil {
		t.Fatalf("optional config must forgive true absence: %v", err)
	}

	cfgEquals(t, got, loadCfgWant(5, "fill", nestedCfgWant("dh", 9)))

	if len(prov) != 0 {
		t.Fatalf("absence contributes no sources, got provenance %v", prov)
	}

	if got := buf.String(); got != "[warn] No config file found. Using default values.\n" {
		t.Fatalf("log output %q, want the exact warning line once", got)
	}
}

// TestOptionalAbsenceZeroValueLoads pins no config and no defaults: the
// caller opted in, so the zero value loads fine.
func TestOptionalAbsenceZeroValueLoads(t *testing.T) {
	paths := optionalAbsencePaths(t)
	captureWarn(t)

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(paths...), WithOptionalConfig())
	if err != nil {
		t.Fatalf("optional config with no defaults must load the zero value: %v", err)
	}

	cfgEquals(t, got, loadCfg{Region: "", Nested: nestedCfg{Host: "", Port: 0}, Retries: 0})

	if len(prov) != 0 {
		t.Fatalf("absence contributes no sources, got provenance %v", prov)
	}
}
