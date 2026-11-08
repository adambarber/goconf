package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// hit is one fragment a refusal must (hit) or must not (miss) contain.
type hit string

// tagsOnlyCfg is a target declaring one list field; only the parse fault
// on its key matters here.
type tagsOnlyCfg struct {
	Tags []tagWord
}

// TestTOMLKeyWithEscapedQuoteIsNotMangled pins that a parse fault on a key
// carrying an escaped quote keeps the key intact or drops the
// parenthetical — never renders a mangled half-key that swallows the fix.
func TestTOMLKeyWithEscapedQuoteIsNotMangled(t *testing.T) {
	body := SourceBytes("\"a\\\"b\" = zzq\n")

	var want tagsOnlyCfg

	err := decodeTOMLStrict(body, FilePath("app.toml"), Target{Value: reflect.ValueOf(&want).Elem()})
	if err == nil {
		t.Fatalf("a parse fault must be refused")
	}

	got := err.Error()

	// The fix sentence must survive: the escaped quote inside the key
	// must not let the strip cut mid-key and swallow it.
	if !contains(hit(got), "check the last line for a missing bracket or quote") {
		t.Fatalf("refusal %q lost the fix sentence", got)
	}

	// Either the full key survives or the parenthetical is gone entirely;
	// a mangled fragment like `last key bzzq.` is never rendered.
	if contains(hit(got), "last key b") && !contains(hit(got), `last key "a\"b"`) {
		t.Fatalf("refusal %q renders a mangled key", got)
	}
}

// contains reports whether the refusal text carries the fragment.
func contains(haystack, needle hit) bool { return strings.Contains(string(haystack), string(needle)) }
