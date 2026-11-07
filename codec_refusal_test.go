package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// secretShape and secretShort are the two value shapes yaml's terror
// echoes: a long scalar truncated to its first seven characters, and a
// short scalar echoed whole. Both are fake fixture shapes, not secrets.
const (
	//nolint:gosec // fixture shape a third-party decoder would echo, not a credential
	secretShape = "sk-live-9f8e7d6c"
	secretShort = "abcdefg"
)

// echoFragments are value fragments that must never reach refusal output.
type (
	echoFragments []string
	echoCase      struct {
		name   caseName
		decode func(SourceBytes, Target) error
		data   SourceBytes
	}
)

// TestRefusalNeverEchoesValues pins the codec-boundary invariant: refusal
// text names keys, types, and line numbers, never config values, even the
// fragments third-party decoders echo of the offending scalar.
func TestRefusalNeverEchoesValues(t *testing.T) {
	// yaml truncates a long scalar to its first seven characters; the
	// truncation itself is also a fragment of the secret.
	fragments := echoFragments{secretShape, secretShort, "sk-live"}

	cases := []echoCase{
		{
			name:   "yaml long value fragment",
			data:   SourceBytes("region: eu\nretries: " + secretShape + "\n"),
			decode: decodeYAMLStrict,
		},
		{
			name:   "yaml short value whole",
			data:   SourceBytes("region: eu\nretries: " + secretShort + "\n"),
			decode: decodeYAMLStrict,
		},
		{
			name:   "yaml nested value fragment",
			data:   SourceBytes("nested:\n  port: " + secretShape + "\n"),
			decode: decodeYAMLStrict,
		},
		{
			name:   "json wrong-typed value",
			data:   SourceBytes(`{"region": "eu", "retries": "` + secretShape + `"}` + "\n"),
			decode: decodeJSONStrict,
		},
		{
			name:   "toml wrong-typed value",
			data:   SourceBytes("region = \"eu\"\nretries = \"" + secretShape + "\"\n"),
			decode: decodeTOMLStrict,
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			refusal := decodeExpectingRefusal(t, tc.decode, tc.data)
			assertNoValueEcho(t, refusal, fragments)
		})
	}
}

// assertNoValueEcho fails the test if the refusal echoes any value
// fragment or drops the line number the decoder reported.
func assertNoValueEcho(t *testing.T, refusal refusalText, fragments echoFragments) {
	t.Helper()

	for _, fragment := range fragments {
		if strings.Contains(string(refusal), fragment) {
			t.Fatalf("refusal %q leaks value fragment %q", refusal, fragment)
		}
	}

	if !strings.Contains(string(refusal), "line") {
		t.Fatalf("refusal %q must still name the line", refusal)
	}
}

// decodeExpectingRefusal runs one strict decode and fails the test unless
// it is refused; the refusal text is returned for fragment assertions.
func decodeExpectingRefusal(t *testing.T, decode func(SourceBytes, Target) error, data SourceBytes) refusalText {
	t.Helper()

	var got codecCfg

	err := decode(data, Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatalf("expected a refusal for wrong-typed value")
	}

	return refusalText(err.Error())
}
