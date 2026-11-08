package goconf

import (
	"errors"
	"strings"

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

// tomlLastKeyHead opens the library's own key-naming parenthetical; the
// quoted span inside it is a key, never a config value.
const tomlLastKeyHead = `(last key "`

// tomlLastKeyTail closes it.
const tomlLastKeyTail = `")`

// stripTOMLQuotedSpans removes every double-quoted span of one message
// except the key in the library's '(last key "K")' parenthetical — that
// quoting is the library's own — and drops a dangling empty one.
func stripTOMLQuotedSpans(msg tomlMessage) tomlMessage {
	head, rest, ok := strings.Cut(string(msg), tomlLastKeyHead)
	if !ok {
		return tomlMessage(stripQuotedSpans(refusalText(msg), '"'))
	}

	key, tail, _ := strings.Cut(rest, tomlLastKeyTail)
	if strings.Contains(key, `"`) {
		return tomlMessage(stripQuotedSpans(refusalText(msg), '"'))
	}

	stripped := stripQuotedSpans(refusalText(tail), '"')

	return tomlMessage(head + tomlLastKeyHead + key + tomlLastKeyTail + string(stripped))
}

// tomlMessage is one BurntSushi parse refusal under rewrite.
type tomlMessage string

// spanText is one span of text inside a decoder message.
type spanText string

// keyQuoted is one key rendered in quotes inside a refusal.
type keyQuoted string

// unquotedKey is one key with its rendering quotes stripped.
type unquotedKey string

// unquote strips the rendering quotes from one quoted key.
func (k keyQuoted) unquote() unquotedKey { return unquotedKey(strings.Trim(string(k), `"`)) }
