package goconf

import (
	"strings"
	"testing"
)

// TestNotFoundProseSpeaksToTheUser pins the not-found refusal shape.
func TestNotFoundProseSpeaksToTheUser(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	t.Chdir(string(root))

	_, err := discover(Discovery{Search: []DirPath{DirPath(string(root))}, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("empty discovery must be refused")
	}

	for _, want := range []userSpan{
		"no config file was found",
		"the app looked for app.json in:",
		"create your config at one of those locations",
	} {
		if !strings.Contains(err.Error(), string(want)) {
			t.Fatalf("refusal %q must contain %q", err, want)
		}
	}

	for _, ban := range []string{"widen the search", "some format spells"} {
		if strings.Contains(err.Error(), ban) {
			t.Fatalf("refusal %q must not contain %q", err, ban)
		}
	}
}
