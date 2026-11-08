package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// TestTOMLScalarIntoListSpeaksOfAList pins the one value vocabulary across
// formats: a scalar written where the target declares a list is a
// "not a list" mismatch in TOML like in JSON and YAML.
func TestTOMLScalarIntoListSpeaksOfAList(t *testing.T) {
	var got listTagCfg

	err := decodeTOMLStrict(SourceBytes("tags = \"x\"\n"), FilePath("app.toml"),
		Target{Value: reflect.ValueOf(&got).Elem()})
	if err == nil {
		t.Fatalf("scalar into a list-typed field must be refused")
	}

	for _, want := range []string{"is text, not a list", `change the value of "tags" to a list`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal %q must contain %q", err, want)
		}
	}

	if strings.Contains(err.Error(), "not a table") {
		t.Fatalf("refusal %q names the wrong shape", err)
	}
}
