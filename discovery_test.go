package goconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test directory names shared across the discovery tests.
const (
	dirGlobal = searchDir("global")
	dirLocal  = searchDir("local")
)

// searchDir is one directory searched for config files.
type searchDir string

// Test file names shared across the discovery tests.
const (
	fileNameJSON = FileName("app.json")
	fileNameYAML = FileName("app.yaml")
	fileNameTOML = FileName("app.toml")
)

// writeTree creates files under a fresh temp root and returns the root.
func writeTree(t *testing.T, files map[searchDir]map[FileName]SourceBytes) DirPath {
	t.Helper()

	root := t.TempDir()

	for dir, entries := range files {
		full := DirPath(filepath.Join(root, string(dir)))
		if err := os.MkdirAll(string(full), 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}

		for name, data := range entries {
			writeFileAt(t, full, name, data)
		}
	}

	return DirPath(root)
}

func TestDiscoverGlobalThenLocal(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
		dirLocal:  {fileNameJSON: SourceBytes(`{"region":"us"}`)},
	})
	paths := legPaths(root, dirGlobal, dirLocal)

	found := discoverOK(t, Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})

	if len(found) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(found), found)
	}

	if found[0].Path != FilePath(filepath.Join(string(root), "global", string(fileNameJSON))) {
		t.Errorf("global file must come first: %v", found[0].Path)
	}
}

func TestDiscoverFormatNarrowsSearch(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu"}`), fileNameYAML: SourceBytes("region: eu\n")},
	})
	paths := legPaths(root, dirGlobal)

	found := discoverOK(t, Discovery{Search: paths, Name: configName("app"), Format: FormatYAML})

	if len(found) != 1 || found[0].Path != FilePath(filepath.Join(string(root), "global", string(fileNameYAML))) {
		t.Fatalf("got %v, want only app.yaml", found)
	}
}

func TestDiscoverRefusesAmbiguousDirectory(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{}`), fileNameTOML: SourceBytes("")},
	})
	paths := legPaths(root, dirGlobal)

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: undetermined})

	mustRefuseNames(t, err, string(root), string(fileNameJSON), string(fileNameTOML))
}

func TestDiscoverRefusesInventedFormat(t *testing.T) {
	_, err := discover(Discovery{Search: []DirPath{"."}, Name: configName("app"), Format: Format(".yml")})

	if err == nil || !strings.Contains(err.Error(), ".yml") || !strings.Contains(err.Error(), acceptedFormats) {
		t.Fatalf("expected refusal naming .yml and accepted spellings, got %v", err)
	}
}

// testText is text inside a discovery test assertion.
type testText string

// escapeCase is one hostile config name discovery must refuse.
type escapeCase struct {
	name configName
	hint testText
}

func TestDiscoverRefusesPathEscapingConfigNames(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	paths := legPaths(root, "search")

	cases := []escapeCase{
		{name: configName("../outside-cfg"), hint: "../outside-cfg"},
		{name: configName("/etc/app"), hint: "/etc/app"},
		{name: configName("a/b"), hint: "a/b"},
		{name: configName("..\\x"), hint: "..\\\\x"},
	}

	for _, tc := range cases {
		_, err := discover(Discovery{Search: paths, Name: tc.name, Format: FormatJSON})
		if err == nil {
			t.Errorf("name %q must be refused", tc.name)
		} else if !strings.Contains(err.Error(), string(tc.hint)) {
			t.Errorf("refusal %q must name the offending name %q", err, tc.hint)
		}
	}
}

func TestDiscoverValidNamesStillBuildCandidates(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"search": {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
	})
	paths := legPaths(root, "search")

	found := discoverOK(t, Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})

	if len(found) != 1 || found[0].Path != FilePath(filepath.Join(string(root), "search", string(fileNameJSON))) {
		t.Fatalf("valid name must still locate candidates, got %v", found)
	}
}

func TestLoadCannotEscapeSearchPathsByName(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})

	outside := writeFileAt(t, root, "outside-cfg.json", SourceBytes(`{"region":"SECRET"}`))

	_, _, err := Load[loadCfg](Spec{Name: configName("../outside-cfg"), Format: FormatJSON},
		WithSearchPaths(DirPath(filepath.Join(string(root), "search"))))
	if err == nil {
		t.Fatalf("path escape via config name must be refused")
	}

	if !strings.Contains(err.Error(), string(outside)) && !strings.Contains(err.Error(), "../outside-cfg") {
		t.Errorf("refusal %q must name the escape", err)
	}
}
