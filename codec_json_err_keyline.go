package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"io"
)

// jsonKeyLineFrame tracks one open container while searching for a key line.
type jsonKeyLineFrame struct {
	isSeq bool
}

// jsonKeyLine walks the token stream and returns the line of the first
// occurrence of the key in key position; 0 when the walk cannot place it.
// The walk is nesting-aware so keys inside array elements are found.
func jsonKeyLine(data SourceBytes, key ObjectKey) int {
	dec := jsontext.NewDecoder(bytes.NewReader(data), jsontext.AllowDuplicateNames(true))

	stack := []jsonKeyLineFrame{{isSeq: false}}
	expectVal := false

	for {
		tok, err := dec.ReadToken()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0
			}

			return 0
		}

		line := lineAtByte(data, dec.InputOffset())
		if found, hit := jsonKeyHit(tok, key, expectVal, stack, line); hit {
			return found
		}

		expectVal = advanceKeyWalk(tok, expectVal, &stack)
	}
}

// jsonKeyHit reports whether this token names the target key and its line.
func jsonKeyHit(tok jsontext.Token, key ObjectKey, expectVal bool, stack []jsonKeyLineFrame, line int) (int, bool) {
	if expectVal || len(stack) == 0 || stack[len(stack)-1].isSeq || tok.Kind() != '"' {
		return line, false
	}

	return line, ObjectKey(tok.String()) == key
}

// advanceKeyWalk folds one token into the nesting-aware key-line walk.
func advanceKeyWalk(tok jsontext.Token, expectVal bool, stack *[]jsonKeyLineFrame) bool {
	kind := tok.Kind()

	if isKeyLineOpen(kind) {
		*stack = append(*stack, jsonKeyLineFrame{isSeq: kind == '['})
		return false
	}

	if isKeyLineClose(kind) {
		if len(*stack) > 1 {
			*stack = (*stack)[:len(*stack)-1]
		}

		return false
	}

	if expectVal {
		return false
	}

	if kind == '"' && len(*stack) > 0 && !(*stack)[len(*stack)-1].isSeq {
		return true
	}

	return false
}

func isKeyLineOpen(kind jsontext.Kind) bool {
	return kind == '{' || kind == '['
}

func isKeyLineClose(kind jsontext.Kind) bool {
	return kind == '}' || kind == ']'
}
