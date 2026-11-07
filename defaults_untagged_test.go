package goconf

import (
	"testing"
)

// untaggedIn is the leaf the untagged-TOML fixtures decode into: no tags.
type untaggedIn struct {
	V leafValue
}

// untaggedOuter wraps untaggedIn the way the repro target does.
type untaggedOuter struct {
	In untaggedIn
}

// TestDefaultsRespectUntaggedTOMLKeySpellings pins the untagged arm of
// the presence vocabulary: the TOML key walk records the canonical
// spelling, so an explicitly-set value is never overridden by a default.
func TestDefaultsRespectUntaggedTOMLKeySpellings(t *testing.T) {
	spellings := []FieldPath{"in.V", "in.v"}

	for _, spelling := range spellings {
		t.Run(string(spelling), func(t *testing.T) {
			root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
				dirGlobal: {fileNameTOML: SourceBytes(string(spelling) + " = \"wins\"\n")},
			})

			got, _, err := Load[untaggedOuter](Spec{Name: configName("app"), Format: FormatTOML},
				append(loadOpts(globalPathOf(root)), WithDefaults(untaggedOuter{In: untaggedIn{V: "DEFAULT"}}))...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.In.V != "wins" {
				t.Fatalf("explicit value lost to default: got %q", got.In.V)
			}
		})
	}
}

// TestUntaggedTOMLUnknownKeyStillRefused pins strictness: a key under
// the untagged parent that no Go field matches must still be refused.
func TestUntaggedTOMLUnknownKeyStillRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("in.v2 = \"wins\"\n")},
	})

	mustRefuseNames(t, loadRefused(t, Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...), "v2")
}
