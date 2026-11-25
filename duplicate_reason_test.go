package goconf

import (
	"errors"
	"testing"
)

type duplicateName string

type duplicateReasonCfg struct {
	Name duplicateName `json:"name" toml:"name" yaml:"name"`
}

func TestDuplicateKeysCarryDuplicateReasonInEveryFormat(t *testing.T) {
	cases := map[string]struct {
		f    Format
		body SourceBytes
	}{
		"yaml": {f: FormatYAML, body: SourceBytes("name: a\nname: b\n")},
		"toml": {f: FormatTOML, body: SourceBytes("name = \"a\"\nname = \"b\"\n")},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := LoadBytes[duplicateReasonCfg](tc.body, Spec{Name: configName("app"), Format: tc.f})
			if !errors.Is(err, ErrDuplicateKey) {
				t.Fatalf("err = %v, want ErrDuplicateKey", err)
			}
		})
	}
}
