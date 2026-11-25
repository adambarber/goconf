package goconf

import (
	"strings"
	"testing"
)

// TestMultiDocumentYAMLIsRefused pins the refused-not-silently-resolved
// contract through the public surface: a stream of more than one YAML
// document is refused, naming the line of the separator.
func TestMultiDocumentYAMLIsRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("region: eu\n---\nregion: us\n")},
	})

	_, _, err := Load[codecCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err == nil {
		t.Fatalf("multi-document stream must be refused")
	}

	mustRefuseNames(t, err, "app.yaml:2:")
}

// TestMultiDocumentLabelsTheSeparatorOnce pins that the separator's line
// is named once: the frame carries "app.yaml:2", so the Subject does not
// restate the line.
func TestMultiDocumentLabelsTheSeparatorOnce(t *testing.T) {
	err := multiDocRefusal("app.yaml", 2)

	got := err.Error()

	if !strings.Contains(got, "app.yaml:2:") {
		t.Fatalf("refusal %q must carry the frame with the line", got)
	}

	if strings.Contains(got, "starts at line") {
		t.Fatalf("refusal %q restates the separator line the frame already carries", got)
	}

	for _, want := range []string{"more than one document", `delete the "---" line`} {
		if !strings.Contains(got, want) {
			t.Fatalf("refusal %q must contain %q", got, want)
		}
	}
}

// TestSingleDocumentYAMLStillDecodes guards the refusal against
// over-refusal: one document is not a multi-document stream.
func TestSingleDocumentYAMLStillDecodes(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("region: eu\nretries: 3\n")},
	})

	got, _, meta, err := LoadWithMeta[codecCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(meta.BySchemaPath("region")) == 0 {
		t.Fatalf("presence missing: %#v", meta.Entries())
	}

	if got.Region != "eu" || got.Retries != 3 {
		t.Fatalf("got %+v", got)
	}
}
