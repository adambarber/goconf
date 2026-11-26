package goconf

import "testing"

// TestJSONExtrasUnexpectedTokenRefusal drives the extras decoder into an
// unexpected end token where a value was required.
func TestJSONExtrasUnexpectedTokenRefusal(t *testing.T) {
	for _, src := range []SourceBytes{
		SourceBytes(`{"resource":}`),
		SourceBytes(`{"resource":[}`),
		SourceBytes(`{"resource":{"name":}}`),
		SourceBytes(`{"resource":`),
		SourceBytes(`{"resource":{"name":`),
		SourceBytes(`{"resource":[`),
		SourceBytes(`{"resource":{"name":"a"`),
		SourceBytes(`{"resource":{1:2}}`),
	} {
		_, _, err := LoadBytes[fuzzExtrasCfg](src, Spec{Name: configName("app"), Format: FormatJSON})
		if err == nil {
			t.Fatalf("src %q loaded, want refusal", src)
		}
	}
}
