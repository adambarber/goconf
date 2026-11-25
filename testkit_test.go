package goconf

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeFileAt writes one file under dir and returns its full path.
func writeFileAt(t *testing.T, dir DirPath, name FileName, data SourceBytes) FilePath {
	t.Helper()

	full := FilePath(filepath.Join(string(dir), string(name)))
	if err := os.WriteFile(string(full), data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	return full
}

// legPaths builds the search paths for the named legs of the temp tree.
func legPaths(root DirPath, dirs ...searchDir) []DirPath {
	paths := make([]DirPath, 0, len(dirs))
	for _, dir := range dirs {
		paths = append(paths, DirPath(filepath.Join(string(root), string(dir))))
	}

	return paths
}

// chdirLocal chdirs into the tree's local search leg.
func chdirLocal(t *testing.T, root DirPath) {
	t.Helper()

	t.Chdir(filepath.Join(string(root), string(dirLocal)))
}

// oneLegTree roots a temp tree whose local leg holds the given files.
func oneLegTree(t *testing.T, files map[FileName]SourceBytes) DirPath {
	t.Helper()

	return writeTree(t, map[searchDir]map[FileName]SourceBytes{dirLocal: files})
}

// symlinkCandidate links name in the local leg at target.
func symlinkCandidate(t *testing.T, root DirPath, name FileName, target FilePath) {
	t.Helper()

	link := filepath.Join(string(root), string(dirLocal), string(name))
	if err := os.Symlink(string(target), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
}

// stripAccess denies access to path and restores the original bits at
// cleanup, so temp-dir removal can still traverse and unlink.
func stripAccess(t *testing.T, path FilePath, deny, restore os.FileMode) {
	t.Helper()

	if err := os.Chmod(string(path), deny); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(string(path), restore); err != nil {
			t.Logf("cleanup chmod: %v", err)
		}
	})
}

// mustRefuseNames asserts err is a refusal naming every required fragment.
func mustRefuseNames(t *testing.T, err error, needs ...string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected a refusal, got nil")
	}

	for _, need := range needs {
		if !strings.Contains(err.Error(), need) {
			t.Errorf("refusal %q must name %q", err, need)
		}
	}
}

// loadOK runs one Load[loadCfg] and fails the test on any error.
func loadOK(t *testing.T, spec Spec, opts ...Option) (loadCfg, []Provenance) {
	t.Helper()

	got, prov, err := Load[loadCfg](spec, opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return got, prov
}

// loadRefused runs one Load[loadCfg] and fails unless it is refused.
func loadRefused(t *testing.T, spec Spec, opts ...Option) error {
	t.Helper()

	_, _, err := Load[loadCfg](spec, opts...)
	if err == nil {
		t.Fatalf("load must be refused")
	}

	return err
}

// refuseLoadCfgBytes loads one in-memory source through the public API and
// fails unless it is refused. Single-source format tests use this instead of
// manufacturing a directory tree and chdir.
func refuseLoadCfgBytes(t *testing.T, data SourceBytes, format Format, want WantText) error {
	t.Helper()

	err := refuseBytesErr[loadCfg](t, data, format)
	if err == nil {
		t.Fatalf("%s", want)
	}

	return err
}

// loadResourceOK loads one in-memory source into the shared extras fixture.
func loadResourceOK(t *testing.T, data SourceBytes, format Format) resourceDoc {
	t.Helper()

	got, _, err := LoadBytes[resourceDoc]([]byte(data), Spec{Name: configName("app"), Format: format})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return got
}

// assertSchemaPaths pins that each named schema path has an occurrence.
func assertSchemaPaths(t *testing.T, meta Meta, paths ...SchemaPath) {
	t.Helper()

	for _, p := range paths {
		if len(meta.BySchemaPath(p)) == 0 {
			t.Errorf("schema path %q missing: %#v", p, meta.Entries())
		}
	}
}

// assertSchemaAbsent pins that no occurrence was recorded for a path.
func assertSchemaAbsent(t *testing.T, meta Meta, paths ...SchemaPath) {
	t.Helper()

	for _, p := range paths {
		if len(meta.BySchemaPath(p)) != 0 {
			t.Errorf("schema path %q must be absent: %#v", p, meta.Entries())
		}
	}
}

// assertOccurrencePaths pins that each named indexed path has an occurrence.
func assertOccurrencePaths(t *testing.T, meta Meta, paths ...OccurrencePath) {
	t.Helper()

	for _, p := range paths {
		if len(meta.ByPath(p)) != 1 {
			t.Errorf("occurrence path %q missing: %#v", p, meta.Entries())
		}
	}
}

// cfgEquals asserts got equals the fully-spelled expected value.
func cfgEquals(t *testing.T, got, want loadCfg) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// defaultOpts layers WithDefaults on top of the common load options.
func defaultOpts(root DirPath, d loadCfg) []Option {
	return append(loadOpts(globalPathOf(root)), WithDefaults(d))
}

// loadDefaultsOK loads from a tree whose config is in the global leg,
// with defaults applied, from the tree root; fails the test on error.
func loadDefaultsOK(t *testing.T, root DirPath, format Format, d loadCfg) loadCfg {
	t.Helper()

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: format},
		defaultOpts(root, d)...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return got
}
