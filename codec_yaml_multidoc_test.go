package goconf

import (
	"reflect"
	"testing"
)

// TestMultiDocumentYAMLIsRefused pins the refused-not-silently-resolved
// contract: a stream of more than one YAML document is refused, naming
// the line of the separator, in both the decode and the presence walk.
func TestMultiDocumentYAMLIsRefused(t *testing.T) {
	multi := SourceBytes("region: eu\n---\nregion: us\n")

	var got codecCfg

	err := decodeYAMLStrict(multi, FilePath(string(fileNameYAML)), Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatalf("multi-document stream must be refused")
	}

	mustRefuseNames(t, err, "app.yaml:2:")

	_, walkErr := presentYAMLKeys(multi, FilePath(string(fileNameYAML)))
	if walkErr == nil {
		t.Fatalf("multi-document stream must be refused in the presence walk")
	}

	mustRefuseNames(t, walkErr, "app.yaml:2:")
}

// TestSingleDocumentYAMLStillDecodes guards the refusal against
// over-refusal: one document, with or without a leading separator, is
// not a multi-document stream.
func TestSingleDocumentYAMLStillDecodes(t *testing.T) {
	single := SourceBytes("region: eu\nretries: 3\n")

	keys, err := presentYAMLKeys(single, FilePath(string(fileNameYAML)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keysPresent(t, keys, []FieldPath{fpRegion}, nil)

	var got codecCfg

	if err := decodeYAMLStrict(single, FilePath(string(fileNameYAML)), Target{Value: reflect.ValueOf(&got).Elem()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Region != "eu" || got.Retries != 3 {
		t.Fatalf("got %+v", got)
	}
}
