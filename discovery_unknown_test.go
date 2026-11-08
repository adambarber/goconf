package goconf

import (
	"strings"
	"testing"
)

// TestLoadRefusesUnknownCandidateExtensionByName: a file matching the
// config name but bearing an extension no format spells must be refused
// naming that file, never skipped into "no config found".
func TestLoadRefusesUnknownCandidateExtensionByName(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {FileName("myconf.json.bak"): SourceBytes(`{}`)},
	})
	paths := legPaths(root, dirLocal)

	_, err := discover(Discovery{Search: paths, Name: configName("myconf"), Format: undetermined})
	if err == nil {
		t.Fatalf("unknown-candidate extension must be refused")
	}

	for _, need := range []string{`".bak" is not a format the app can read`, "delete or rename", "myconf.json.bak"} {
		if !strings.Contains(err.Error(), need) {
			t.Errorf("refusal %q must name %q", err, need)
		}
	}

	if strings.Contains(err.Error(), "no config file found") {
		t.Errorf("refusal %q must not be the empty-discovery refusal", err)
	}
}
