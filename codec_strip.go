package goconf

// The shared value-freedom pass: config values never appear in refusals.

import "strings"

// stripQuotedSpans removes every span between the given quote byte of
// one message; an unbalanced quote drops the suspected tail too.
func stripQuotedSpans(msg refusalText, quote byte) refusalText {
	if !strings.ContainsRune(string(msg), rune(quote)) {
		return msg
	}

	var b strings.Builder

	text := string(msg)

	for {
		open := strings.IndexByte(text, quote)
		if open < 0 {
			break
		}

		b.WriteString(text[:open])

		text = text[open+1:]

		closing := strings.IndexByte(text, quote)
		if closing < 0 {
			return refusalText(strings.TrimSuffix(b.String(), " "))
		}

		text = text[closing+1:]
	}

	b.WriteString(text)

	return refusalText(b.String())
}
