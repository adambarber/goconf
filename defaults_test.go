package goconf

import (
	"path/filepath"
	"testing"
)

// loadDeepCase is one deep default-fill scenario: a source setting part of
// a nested object, defaults filling the absent leaves.
type loadDeepCase struct {
	name     caseName
	fileName FileName
	format   Format
	src      SourceBytes
}

func TestLoadDefaultsFillDeepAbsences(t *testing.T) {
	cases := []loadDeepCase{
		{
			name:     "json: nested leaf set, sibling leaf filled",
			src:      SourceBytes(`{"nested":{"port":3}}`),
			fileName: fileNameJSON,
			format:   FormatJSON,
		},
		{
			name:     "toml: nested leaf set, sibling leaf filled",
			src:      SourceBytes("[nested]\nport = 3\n"),
			fileName: fileNameTOML,
			format:   FormatTOML,
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
				dirLocal: {tc.fileName: tc.src},
			})
			t.Chdir(filepath.Join(string(root), string(dirLocal)))

			got, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: tc.format},
				WithSearchPaths(DirPath(filepath.Join(string(root), string(dirLocal)))),
				WithDefaults(loadCfgWant(0, "fill", nestedCfgWant("DEF", 0))))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := loadCfgWant(0, "fill", nestedCfgWant("DEF", 3))
			if got != want {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		})
	}
}
