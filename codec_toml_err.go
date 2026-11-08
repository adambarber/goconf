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
// except the key in the library's '(last key "K")' parenthetical; an
// ambiguous key shape drops the parenthetical entirely.
func stripTOMLQuotedSpans(msg tomlMessage) tomlMessage {
	head, rest, ok := strings.Cut(string(msg), tomlLastKeyHead)
	if !ok {
		return tomlMessage(stripQuotedSpans(refusalText(msg), '"'))
	}

	key, tail, found := cutKey(tomlMessage(rest))
	if !found || strings.Contains(string(key), `"`) {
		return tomlMessage(head + string(stripQuotedSpans(refusalText(rest), '"')))
	}

	stripped := stripQuotedSpans(refusalText(tail), '"')

	return tomlMessage(head + tomlLastKeyHead + string(key) + tomlLastKeyTail + string(stripped))
}

// cutKey cuts one quoted key at the library's closing parenthetical,
// escape-aware: a backslash-escaped quote inside the key is content, not
// the closing quote.
func cutKey(rest tomlMessage) (keyQuoted, tomlMessage, bool) {
	for i := 0; i < len(rest); i++ {
		if rest[i] == '\\' && i+1 < len(rest) {
			i++
			continue
		}

		if strings.HasPrefix(string(rest[i:]), tomlLastKeyTail) {
			tail := rest[i+len(tomlLastKeyTail):]

			return keyQuoted(rest[:i]), tail, true
		}
	}

	return "", "", false
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
