package goconf

import (
	"errors"
	"strings"
	"testing"
)

type jsonRendererCase string

// TestJSONRefusalRenderersAreValueFree pins the defensive refusal renderers
// that ordinary document paths do not reach. They are part of the value-freedom
// contract, so their output is checked directly.
func TestJSONRefusalRenderersAreValueFree(t *testing.T) {
	cases := map[jsonRendererCase]error{
		"empty":      jsonEmptyRefusal(FilePath("app.json")),
		"unexpected": jsonUnexpectedToken(FilePath("app.json")),
	}

	for name, err := range cases {
		t.Run(string(name), func(t *testing.T) {
			if err == nil {
				t.Fatal("renderer returned no refusal")
			}

			if !strings.Contains(err.Error(), "app.json") {
				t.Fatalf("refusal = %q, want the file named", err)
			}
		})
	}

	if !errors.Is(cases["unexpected"], ErrUnexpectedJSONToken) {
		t.Fatalf("unexpected-token refusal lost its reason: %v", cases["unexpected"])
	}
}

// TestJSONQuotedNameFallback pins the quoted-name extraction used when a
// semantic error carries no usable JSON pointer.
func TestJSONQuotedNameFallback(t *testing.T) {
	got := jsonErrQuotedName(refusalText(`unknown object member name "token"`))
	if got != ObjectKey("token") {
		t.Fatalf("quoted name = %q, want token", got)
	}
}
