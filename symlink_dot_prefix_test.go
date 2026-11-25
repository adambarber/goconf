package goconf

import (
	"path/filepath"
	"testing"
)

type symlinkDotPrefixCfg struct {
	Port int `json:"port"`
}

func TestSymlinkTargetWithDotPrefixLoads(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		"..target.json": SourceBytes(`{"port":7}`),
	})
	local := legPaths(root, dirLocal)[0]
	target := FilePath(filepath.Join(string(local), "..target.json"))
	symlinkCandidate(t, root, FileName("app.json"), target)

	got, _, err := Load[symlinkDotPrefixCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(local))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Port != 7 {
		t.Fatalf("Port = %d, want 7", got.Port)
	}
}
