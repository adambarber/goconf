package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Load-level pins for format spelling: the undetermined format decodes
// each discovered file through its own format, and the refusals around
// candidate formats still fire.
func TestLoadUndeterminedFormatUsesPerFileFormat(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu","nested":{"host":"h","port":1}}`)},
	})
	t.Chdir(string(root))

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: undetermined}, loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := loadCfgWant(0, "eu", nestedCfgWant("h", 1))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	if len(prov) != 1 || prov[0].Path != FilePath(filepath.Join(string(root), string(dirGlobal), string(fileNameJSON))) {
		t.Fatalf("provenance must name the global json file: %v", prov)
	}
}

func TestLoadMixedFormatLegsDecodePerFile(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu","nested":{"host":"h","port":1}}`)},
		dirLocal:  {fileNameTOML: SourceBytes("region = \"us\"\n")},
	})
	t.Chdir(filepath.Join(string(root), string(dirLocal)))

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: undetermined},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// local toml overrides global json's region; the rest comes from json.
	want := loadCfgWant(0, "us", nestedCfgWant("h", 1))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadUndeterminedAmbiguityStillRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {fileNameJSON: SourceBytes(`{"region":"eu"}`), fileNameTOML: SourceBytes("region = \"us\"\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: undetermined})
	if err == nil {
		t.Fatalf("two candidates in one dir with an undetermined format must refuse")
	}
}

func TestLoadUnknownCandidateExtensionStillRefused(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.bak"), []byte(`{"region":"eu"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Chdir(root)

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: undetermined})
	if !errors.Is(err, errUnknownCandidate) {
		t.Fatalf("got %v, want errUnknownCandidate", err)
	}
}
