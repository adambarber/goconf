package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// secretFragment is a secret-shaped scalar planted in error positions.
const secretFragment = "zzq-secret-marker-8842"

// keyIsSafe is a key-name refusal that legitimately quotes keys: the
// strip pass must not mangle these.
const keyIsSafe = `unknown field: "safekey"`

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
			decode: fileDecode(decodeTOMLStrict, fileNameTOML),
		},
		{
			name:   "toml invalid escape echo",
			data:   SourceBytes(`retries = "bad\` + secretFragment + `"` + "\n"),
			decode: fileDecode(decodeTOMLStrict, fileNameTOML),
		},
		{
			name:   "toml type mismatch",
			data:   SourceBytes("retries = \"" + secretFragment + "\"\n"),
			decode: fileDecode(decodeTOMLStrict, fileNameTOML),
		},
		{
			name:   "json type mismatch",
			data:   SourceBytes(`{"retries": "` + secretFragment + `"}`),
			decode: fileDecode(decodeJSONStrict, fileNameJSON),
		},
		{
			name:   "json syntax error near value",
			data:   SourceBytes("{\n  \"retries\": " + secretFragment + "\n}"),
			decode: fileDecode(decodeJSONStrict, fileNameJSON),
		},
		{
			name:   "yaml type error",
			data:   SourceBytes("retries: " + secretFragment + "\n"),
			decode: fileDecode(decodeYAMLStrict, fileNameYAML),
		},
		{
			name:   "yaml backquoted value",
			data:   SourceBytes("retries: `x`" + secretFragment + "y`\n"),
			decode: fileDecode(decodeYAMLStrict, fileNameYAML),
		},
		{
			name:   "yaml scanner error on value line",
			data:   SourceBytes("region: eu\nretries: @x`" + secretFragment + "\n"),
			decode: fileDecode(decodeYAMLStrict, fileNameYAML),
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
			assertNoValueEcho(t, refusal, echoFragments{secretFragment, "zzq", "marker-8842"})
		})
	}
}

// TestKeyNamingRefusalsStillQuoteKeys pins the other half of the
// invariant: keys are safe by contract, so the strip pass must leave
// key-naming refusals intact.
func TestKeyNamingRefusalsStillQuoteKeys(t *testing.T) {
	var got codecCfg

	err := fileDecode(decodeJSONStrict, fileNameJSON)(SourceBytes(`{"region":"eu","safekey":1}`), Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	if !strings.Contains(err.Error(), keyIsSafe) {
		t.Fatalf("key-naming refusal must keep the quoted key, got %q", err)
	}
}

// TestNoValueFragmentSurvivesTheTOMLPresenceWalk pins the presence-walk
// half of the invariant: presentTOMLKeys runs before decode, so its parse
// errors must be sanitized like the decode path's.
func TestNoValueFragmentSurvivesTheTOMLPresenceWalk(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("retries = " + secretFragment + "\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("parse-error TOML must be refused")
	}

	refusal := refusalText(err.Error())
	assertNoValueEcho(t, refusal, echoFragments{secretFragment, "zzq", "marker-8842"})
}
