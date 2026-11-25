package goconf

import (
	"errors"
	"testing"

	"github.com/adambarber/gorefuse"
)

// TestRefusalCarriesErrThroughErrorsAs pins the public error-identity
// contract: every refusal a Load returns is a gorefuse.Because[Reason] a
// caller can extract with errors.As and switch on with errors.Is.
func TestRefusalCarriesErrThroughErrorsAs(t *testing.T) {
	err := refuseLoadCfgBytes(t, SourceBytes(`{"region":"eu","region":"us"}`), FormatJSON, WantText("duplicate key must be refused"))

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
	_, _, err := LoadBytes[loadCfg]([]byte("{\n  \"region\": \"eu\",\n  \"extra\": 1\n}"),
		Spec{Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	want := FilePath(configName("app"))
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
	err, _ := emptyDiscoveryRefusal(t, dirGlobal, dirLocal)

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
		ErrIO, ErrUnknownField, ErrTypeMismatch,
		ErrNothingFound, ErrUnmarshalable, ErrBadDefaultShape,
		ErrTargetCast, ErrSourceTooLarge, ErrNeedDeclaration, ErrEnvParse, ErrEnvAmbiguous,
		ErrUnsupportedValueContainer,
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

// TestTOMLUnknownFieldCarriesErrUnknownField pins that a TOML key the
// target does not declare refuses with ErrUnknownField, in the presence
// walk (the stage that refuses before decode) and in decode alike.
func TestTOMLUnknownFieldCarriesErrUnknownField(t *testing.T) {
	err := refuseLoadCfgBytes(t, SourceBytes("region = \"eu\"\nnonesuch = 1\n"), FormatTOML, WantText("unknown TOML field must be refused"))

	var because gorefuse.Because[Reason]
	if !errors.As(err, &because) {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	if because.Reason != ErrUnknownField {
		t.Fatalf("reason = %v, want ErrUnknownField (refusal: %s)", because.Reason, err)
	}
}
