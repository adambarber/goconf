package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// secretFragment is a secret-shaped scalar planted in error positions.
//
//nolint:gosec // fixture shape a decoder would echo, not a credential
const secretFragment = "SE-CRET-VALUE"

// keyIsSafe is a key-name refusal that legitimately quotes keys: the
// strip pass must not mangle these.
const keyIsSafe = `unknown field "safekey"`

// valueFreeCase is one cross-format invariant row: data planted with the
// secret fragment, decoded, and its refusal checked for fragments.
type valueFreeCase struct {
	name   caseName
	decode func(SourceBytes, Target) error
	data   SourceBytes
}

// valueFreeCases plants the secret in each format's error positions.
func valueFreeCases() []valueFreeCase {
	return []valueFreeCase{
		{
			name:   "toml expected-value echo",
			data:   SourceBytes("retries = " + secretFragment + "\n"),
			decode: decodeTOMLStrict,
		},
		{
			name:   "toml invalid escape echo",
			data:   SourceBytes(`retries = "bad\` + secretFragment + `"` + "\n"),
			decode: decodeTOMLStrict,
		},
		{
			name:   "toml type mismatch",
			data:   SourceBytes("retries = \"" + secretFragment + "\"\n"),
			decode: decodeTOMLStrict,
		},
		{
			name:   "json type mismatch",
			data:   SourceBytes(`{"retries": "` + secretFragment + `"}`),
			decode: decodeJSONStrict,
		},
		{
			name:   "json syntax error near value",
			data:   SourceBytes("{\n  \"retries\": " + secretFragment + "\n}"),
			decode: decodeJSONStrict,
		},
		{
			name:   "yaml type error",
			data:   SourceBytes("retries: " + secretFragment + "\n"),
			decode: decodeYAMLStrict,
		},
		{
			name:   "yaml backquoted value",
			data:   SourceBytes("retries: `x`" + secretFragment + "y`\n"),
			decode: decodeYAMLStrict,
		},
		{
			name:   "yaml scanner error on value line",
			data:   SourceBytes("region: eu\nretries: @x`" + secretFragment + "\n"),
			decode: decodeYAMLStrict,
		},
	}
}

// TestNoValueFragmentSurvivesAnyRefusal pins the cross-format invariant:
// no config value fragment survives into refusal text, in any format, on
// any error path.
func TestNoValueFragmentSurvivesAnyRefusal(t *testing.T) {
	for _, tc := range valueFreeCases() {
		t.Run(string(tc.name), func(t *testing.T) {
			refusal := decodeExpectingRefusal(t, tc.decode, tc.data)
			assertNoValueEcho(t, refusal, echoFragments{secretFragment, "SE-CRET", "CRET-VAL"})
		})
	}
}

// TestKeyNamingRefusalsStillQuoteKeys pins the other half of the
// invariant: keys are safe by contract, so the strip pass must leave
// key-naming refusals intact.
func TestKeyNamingRefusalsStillQuoteKeys(t *testing.T) {
	var got codecCfg

	err := decodeJSONStrict(SourceBytes(`{"region":"eu","safekey":1}`), Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	if !strings.Contains(err.Error(), keyIsSafe) {
		t.Fatalf("key-naming refusal must keep the quoted key, got %q", err)
	}
}
