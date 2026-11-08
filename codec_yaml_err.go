package goconf

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// errValueFree rewrites a decoder error so no scalar value survives;
// yaml's typed errors carry one message per failure.
func errValueFree(err error) error {
	var typeErr *yaml.TypeError

	if !errors.As(err, &typeErr) {
		return refusalError(stripBackquotedValues(refusalText(err.Error())))
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
