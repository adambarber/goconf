package goconf

import (
	"errors"
	"strings"
	"testing"

	"github.com/adambarber/gorefuse"
)

// TestJSONRootArrayIsNotAMapping pins the one fault vocabulary: a JSON file
// holding a root array is refused with Reason "not a mapping", the same
// fault YAML names, whichever decoder noticed it.
func TestJSONRootArrayIsNotAMapping(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes("[1,2,3]")},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("a root array must be refused")
	}

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	if because.Reason != ErrNotAMapping {
		t.Fatalf("reason = %v, want ErrNotAMapping (refusal: %s)", because.Reason, err)
	}

	for _, want := range []string{"the file holds a list", "settings must be name: value pairs"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal %q must contain %q", err, want)
		}
	}
}
