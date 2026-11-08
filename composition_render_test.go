package goconf

import (
	"path/filepath"
	"testing"
)

// TestCompositionRenderPinsTheOwnerModel pins goconf's half of the
// gorefuse composition contract byte for byte through a real Load: the
// frame:line form — the location rides the frame, the field the reason.
func TestCompositionRenderPinsTheOwnerModel(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"region":"eu","extra":1}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	want := "Error: " + filepath.Join(string(root), string(fileNameJSON)) +
		`:1: unknown field: "extra". remove it from your config file, or the app must declare the field in its schema.`

	if err.Error() != want {
		t.Fatalf("render = %q, want %q", err.Error(), want)
	}
}
