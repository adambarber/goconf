package goconf

import (
	"os"
	"path/filepath"
	"testing"
)

// loadCfgWant composes a fully-spelled expected value.
func loadCfgWant(retries int, reg region, nested nestedCfg) loadCfg {
	return loadCfg{Nested: nested, Region: reg, Retries: retries}
}

// nestedCfgWant composes a fully-spelled nested config.
func nestedCfgWant(h host, port int) nestedCfg {
	return nestedCfg{Host: h, Port: port}
}

type loadCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// loadOpts builds the common options: one global dir, cwd as local.
func loadOpts(global DirPath) []Option {
	opt := WithGlobalDir(global)

	return []Option{opt}
}

// globalPathOf returns the tree's global directory for the temp root.
func globalPathOf(root DirPath) DirPath {
	return DirPath(filepath.Join(string(root), string(dirGlobal)))
}

func TestLoadMergesGlobalThenLocal(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu","nested":{"host":"h","port":1}}`)},
		dirLocal:  {fileNameJSON: SourceBytes(`{"region":"us"}`)},
	})

	t.Chdir(filepath.Join(string(root), string(dirLocal)))

	got, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(DirPath(filepath.Join(string(root), string(dirGlobal))))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := loadCfgWant(0, "us", nestedCfgWant("h", 1))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	if len(prov) != 2 {
		t.Fatalf("got %d provenance entries, want 2: %v", len(prov), prov)
	}

	if prov[0].Path != FilePath(filepath.Join(string(root), string(dirGlobal), string(fileNameJSON))) {
		t.Errorf("first provenance entry must be the global file: %v", prov[0].Path)
	}
}

func TestLoadDefaultsFillOnlyAbsences(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu","retries":0}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		append(loadOpts(globalPathOf(root)), WithDefaults(loadCfgWant(5, "fill", nestedCfgWant("", 9))))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// retries is present with an explicit zero value: the default must not
	// override it. region and nested.port are absent: defaults fill them.
	want := loadCfgWant(0, "eu", nestedCfgWant("", 9))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadExplicitNullIsNeverOverridden(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":null}`)},
	})
	t.Chdir(string(root))

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		append(loadOpts(globalPathOf(root)), WithDefaults(loadCfgWant(0, "fill", nestedCfgWant("", 0))))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Region != "" {
		t.Fatalf("explicit null is a decision and must not be overridden, got %q", got.Region)
	}
}

func TestLoadDefaultsFailsLoudOnUnmarshalableDefault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected a loud failure at option construction")
		}
	}()

	type brokenCfg struct {
		Ch chan int `json:"ch"`
	}

	_ = WithDefaults(brokenCfg{Ch: make(chan int)})
}

func TestLoadSearchPathsOverridesDefault(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"custom": {fileNameJSON: SourceBytes(`{"region":"custom"}`)},
		dirLocal: {fileNameJSON: SourceBytes(`{"region":"local"}`)},
	})
	t.Chdir(filepath.Join(string(root), string(dirLocal)))

	got, prov, err := Load[loadCfg](
		Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(filepath.Join(string(root), "custom"))),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Region != "custom" {
		t.Fatalf("got %q, want custom", got.Region)
	}

	if len(prov) != 1 {
		t.Fatalf("got %d provenance entries, want 1", len(prov))
	}
}

func TestLoadDigestStampsEachSource(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes(`{"region":"eu"}`)},
	})
	t.Chdir(string(root))

	_, prov, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prov) == 0 || len(prov[0].Digest) == 0 {
		t.Fatalf("digest must be stamped on every contributing source: %v", prov)
	}

	raw, readErr := os.ReadFile(string(prov[0].Path))
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}

	if string(raw) != `{"region":"eu"}` {
		t.Fatalf("fixture drifted: %q", raw)
	}
}

func TestLoadYAMLExplicitNullIsNeverOverridden(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("region: null\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatYAML},
		append(loadOpts(globalPathOf(root)), WithDefaults(loadCfgWant(7, "fill", nestedCfgWant("dh", 9))))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Region != "" {
		t.Fatalf("explicit null is a decision and must not be overridden, got %q", got.Region)
	}

	// absent fields still default-fill despite the null sibling.
	want := loadCfgWant(7, "", nestedCfgWant("dh", 9))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
