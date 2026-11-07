package goconf

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// errDecoderRefused is the sentinel for a rewritten refusal: identity
// survives, value text does not.
var errDecoderRefused = errors.New("decoder refused")

// errValueFree rewrites a decoder error so no scalar value survives;
// yaml's typed errors carry one message per failure.
func errValueFree(err error) error {
	var typeErr *yaml.TypeError

	if !errors.As(err, &typeErr) {
		return fmt.Errorf("%w: %s", errDecoderRefused, stripBackquotedValues(refusalText(err.Error())))
	}

	stripped := make([]string, 0, len(typeErr.Errors))

	for _, e := range typeErr.Errors {
		stripped = append(stripped, string(stripYAMLTypeLine(refusalText(e))))
	}

	return &yaml.TypeError{Errors: stripped}
}

// stripBackquotedValues removes every backquoted fragment of a message:
// yaml quotes values, never keys, in backquotes.
func stripBackquotedValues(msg refusalText) refusalText {
	return stripQuotedSpans(msg, '`')
}
