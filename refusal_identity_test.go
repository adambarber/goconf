package goconf

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/adambarber/gorefuse"
)

// TestRefusalCarriesErrThroughErrorsAs pins the public error-identity
// contract: every refusal a Load returns is a gorefuse.Because[Reason] a
// caller can extract with errors.As and switch on with errors.Is.
func TestRefusalCarriesErrThroughErrorsAs(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"region":"eu","region":"us"}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("duplicate key must be refused")
	}

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	if because.Reason != ErrDuplicateKey || !errors.Is(because.Reason, ErrDuplicateKey) {
		t.Fatalf("reason = %v, want ErrDuplicateKey", because.Reason)
	}
}

// TestRefusalFrameNamesTheRealFile pins that a decode refusal carries a
// frame pointing at the file that was refused, not the decoder's internals.
func TestRefusalFrameNamesTheRealFile(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes("{\n  \"region\": \"eu\",\n  \"extra\": 1\n}")},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	want := FilePath(filepath.Join(string(root), string(fileNameJSON)))
	if because.What.File != gorefuse.File(want) {
		t.Fatalf("frame file = %q, want %q", because.What.File, want)
	}

	if because.Reason != ErrUnknownField {
		t.Fatalf("reason = %v, want ErrUnknownField", because.Reason)
	}
}

// TestDiscoveryRefusalIsFrameless pins that a discovery refusal that
// happened before any file exists carries no frame: gorefuse renders no
// file:line prefix for a zero File, and no file is fabricated.
func TestDiscoveryRefusalIsFrameless(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	paths := legPaths(root, dirGlobal, dirLocal)

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("empty discovery must be refused")
	}

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	if because.What.File != "" {
		t.Fatalf("discovery refusal must be frameless, got file %q", because.What.File)
	}

	if because.Reason != ErrNothingFound {
		t.Fatalf("reason = %v, want ErrNothingFound", because.Reason)
	}
}

// TestReasonsAreDistinct pins that every declared reason constant has its
// own rendered spelling: identity is the Reason, not the prose.
func TestReasonsAreDistinct(t *testing.T) {
	seen := map[Reason]bool{}

	for _, r := range []Reason{
		ErrUnknownFormat, ErrAmbiguous, ErrUnknownCandidate,
		ErrEscapingName, ErrOutsideSearch, ErrDuplicateKey,
		ErrMultiDoc, ErrNotAMapping, ErrDecoderRefused,
		ErrKeyUnmappable, ErrUnknownField, ErrTypeMismatch,
		ErrNothingFound, ErrUnmarshalable, ErrBadDefaultShape,
		ErrTargetCast,
	} {
		if r == "" {
			t.Fatalf("a reason constant must not render empty")
		}

		if seen[r] {
			t.Fatalf("reason %q is declared twice", r)
		}

		seen[r] = true
	}
}
