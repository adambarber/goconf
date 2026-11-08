package goconf

import (
	"errors"

	"github.com/BurntSushi/toml"
)

// errTOMLValueFree rewrites a BurntSushi error so no scalar value
// fragment survives; only ParseError renders values. Match the value
// type: parse() recovers ParseError from a panic by value.
func errTOMLValueFree(err error) error {
	var perr toml.ParseError

	if !errors.As(err, &perr) {
		return err
	}

	return refusalError(refusalText(stripTOMLQuotedSpans(tomlMessage(perr.Error()))))
}

// stripTOMLQuotedSpans removes every double-quoted span of one message.
// The value-quoting classes render values through %q, never keys.
func stripTOMLQuotedSpans(msg tomlMessage) tomlMessage {
	return tomlMessage(stripQuotedSpans(refusalText(msg), '"'))
}

// tomlMessage is one BurntSushi parse refusal under rewrite.
type tomlMessage string

// spanText is one span of text inside a decoder message.
type spanText string

// keyQuoted is one key rendered in quotes inside a refusal.
type keyQuoted string
