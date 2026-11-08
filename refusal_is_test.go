package goconf

import (
	"errors"
	"fmt"
	"testing"
)

// TestErrorsIsThroughLoadAndWrap pins the documented public contract: a
// Load refusal answers errors.Is against a Reason constant — directly
// and through a %w wrapper.
func TestErrorsIsThroughLoadAndWrap(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameJSON: SourceBytes(`{"region":"eu","extra":1}`)},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("refusal %v must errors.Is ErrUnknownField", err)
	}

	wrapped := fmt.Errorf("loading config: %w", err)
	if !errors.Is(wrapped, ErrUnknownField) {
		t.Fatalf("wrapped refusal %v must still errors.Is ErrUnknownField", wrapped)
	}

	if errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("refusal %v must not match an unrelated reason", err)
	}
}
