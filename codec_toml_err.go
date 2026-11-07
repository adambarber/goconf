package goconf

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

// errTOMLValueFree rewrites a BurntSushi error so no scalar value
// fragment survives; only ParseError renders values.
func errTOMLValueFree(err error) error {
	var perr *toml.ParseError

	if !errors.As(err, &perr) {
		return err
	}

	return fmt.Errorf("%w: %s", errTOMLRefused,
		stripTOMLQuotedSpans(tomlMessage(perr.Error())))
}

// errTOMLRefused is the sentinel for a rewritten TOML parse refusal:
// identity survives, value text does not.
var errTOMLRefused = errors.New("decoder refused")

// stripTOMLQuotedSpans removes every double-quoted span of one message.
// The value-quoting classes render values through %q, never keys.
func stripTOMLQuotedSpans(msg tomlMessage) tomlMessage {
	return tomlMessage(stripQuotedSpans(refusalText(msg), '"'))
}

// tomlMessage is one BurntSushi parse refusal under rewrite.
type tomlMessage string
