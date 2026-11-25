package goconf

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// scratchWorkDir creates and chdirs into a scratch work directory.
func scratchWorkDir(t *testing.T, scratch DirPath) {
	t.Helper()

	work := filepath.Join(string(scratch), "work")
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	t.Chdir(work)
}

// discoverOK runs one discovery and fails the test on any error.
func discoverOK(t *testing.T, d Discovery) []FoundFile {
	t.Helper()

	found, err := discover(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return found
}

// yamlContentNode parses one YAML document and unwraps its document node.
func yamlContentNode(t *testing.T, data SourceText) *yaml.Node {
	t.Helper()

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		t.Fatal(err)
	}

	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}

	return &doc
}

// nullNotOverridden asserts region stayed empty: an explicit null is a
// decision the loader must respect.
func nullNotOverridden(t *testing.T, got loadCfg) {
	t.Helper()

	if got.Region != "" {
		t.Fatalf("explicit null is a decision and must not be overridden, got %q", got.Region)
	}
}
