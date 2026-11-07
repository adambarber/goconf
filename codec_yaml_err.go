package goconf

// The refusal-text invariant: config values never appear in refusal text.
// Third-party decoder errors are sanitized at this codec boundary — a
// fragment of a secret is still a secret, so no echo survives the wrap.

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// errValueFree rewrites a decoder error so no scalar value fragment
// survives. yaml's typed errors carry one message per decode failure;
// every other error is treated with the same text transform.
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

// errDecoderRefused is the sentinel for a decoder refusal whose raw text
// was value-echoing and has been rewritten; identity survives, text does not.
var errDecoderRefused = errors.New("decoder refused")

// stripBackquotedValues removes every backquoted fragment of a message.
// yaml quotes values, never keys or types, in backquotes; a backquoted
// span is a value by construction, so all of them go.
func stripBackquotedValues(msg refusalText) refusalText {
	text := string(msg)

	if !strings.Contains(text, "`") {
		return msg
	}

	var b strings.Builder

	for {
		open := strings.IndexByte(text, '`')
		if open < 0 {
			break
		}

		b.WriteString(text[:open])

		text = text[open+1:]

		closing := strings.IndexByte(text, '`')
		if closing < 0 {
			// an unbalanced backquote leaves the tail a suspected value
			// fragment: the strip ends at the last close, the tail is dropped
			return refusalText(strings.TrimSuffix(b.String(), " "))
		}

		text = text[closing+1:]
	}

	b.WriteString(text)

	return refusalText(b.String())
}

// terrorHead begins every value-echoing yaml type error line.
const terrorHead = "cannot unmarshal "

// stripYAMLTypeLine rewrites a yaml.TypeError line of the terror shape
// so the echoed scalar never reaches the refusal.
func stripYAMLTypeLine(msg refusalText) refusalText {
	head, tag, ok := strings.Cut(string(msg), terrorHead)
	if !ok {
		return stripBackquotedValues(msg)
	}

	return refusalText(head + terrorHead + string(cutTypeSpan(terrorTail(tag))))
}

// cutTypeSpan drops the echoed scalar from one terror tail: the tag
// token is kept, every backquoted span (balanced or trailing-open) goes,
// and the `into TYPE` clause survives untouched.
func cutTypeSpan(tag terrorTail) terrorTail {
	return terrorTail(stripBackquotedValues(refusalText(tag)))
}

// terrorTail is the part of a yaml terror line after its head.
type terrorTail string
