package goconf

import (
	"errors"
	"strings"
	"testing"

	"github.com/adambarber/gorefuse"
)

// wantFragment is the fragment an expected refusal must name.
type wantFragment string

// yamlVocabCase is one error-identity row.
type yamlVocabCase struct {
	want     Reason
	fragment wantFragment
	name     caseName
	data     SourceBytes
}

// assertYAMLReason fails unless err is a Because[Reason] with tc.want's
// identity naming tc.fragment.
func assertYAMLReason(t *testing.T, tc yamlVocabCase, err error) {
	t.Helper()

	because, ok := errors.AsType[gorefuse.Because[Reason]](err)
	if !ok {
		t.Fatalf("refusal %v must be a gorefuse.Because[Reason]", err)
	}

	if because.Reason != tc.want {
		t.Fatalf("reason = %v, want %v (refusal: %s)", because.Reason, tc.want, err)
	}

	if !strings.Contains(err.Error(), string(tc.fragment)) {
		t.Fatalf("refusal %q must name %q", err, tc.fragment)
	}
}

// TestYAMLUnknownFieldAndTypeMismatchCarryTheSharedVocabulary pins the
// error-identity contract: YAML's unknown-field refusal is
// ErrUnknownField, its type-mismatch refusal is ErrTypeMismatch.
func TestYAMLUnknownFieldAndTypeMismatchCarryTheSharedVocabulary(t *testing.T) {
	cases := []yamlVocabCase{
		{
			name:     "yaml unknown field",
			data:     SourceBytes("region: eu\nnonesuch: 1\n"),
			want:     ErrUnknownField,
			fragment: "nonesuch",
		},
		{
			name:     "yaml type mismatch",
			data:     SourceBytes("retries: not-a-number\n"),
			want:     ErrTypeMismatch,
			fragment: "expected a number, found text",
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			var got codecCfg

			err := decodeYAMLStrict(tc.data, FilePath(string(fileNameYAML)), targetOf(&got))
			if err == nil {
				t.Fatalf("expected refusal")
			}

			assertYAMLReason(t, tc, err)
		})
	}
}
