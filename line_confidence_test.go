package goconf

import "testing"

type lineConfidenceCaseName string

// TestLineConfidenceByFormat pins that metadata declares how authoritative a
// source line is: parser-derived for JSON/YAML, heuristic for TOML.
func TestLineConfidenceByFormat(t *testing.T) {
	cases := []struct {
		name       lineConfidenceCaseName
		format     Format
		data       SourceBytes
		confidence LineConfidence
	}{
		{"json", FormatJSON, SourceBytes("{\n  \"region\": \"eu\",\n  \"port\": 8080\n}"), LineExact},
		{"yaml", FormatYAML, SourceBytes("region: eu\nport: 8080\n"), LineExact},
		{"toml", FormatTOML, SourceBytes("region = \"eu\"\nport = 8080\n"), LineHeuristic},
	}

	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			_, _, meta, err := LoadBytesWithMeta[pathwayConfig](tc.data, Spec{Name: pathwayAppName, Format: tc.format})
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			occurrences := meta.ByPath(OccurrencePath("port"))
			if len(occurrences) != 1 {
				t.Fatalf("occurrences = %+v, want one port", occurrences)
			}

			if occurrences[0].LineConfidence != tc.confidence {
				t.Fatalf("confidence = %v, want %v", occurrences[0].LineConfidence, tc.confidence)
			}
		})
	}
}
