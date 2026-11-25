package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"io"

	"github.com/adambarber/gorefuse"
)

// ObjectKey is one JSON object key exactly as it appears in the document.
type ObjectKey string

// dupContainerKind is the grammar role of one open JSON container.
type dupContainerKind uint8

const (
	dupObject dupContainerKind = iota
	dupArray
)

// dupFrame is one open container: an object carries its declared keys; an
// array carries none, because its elements are values in every position.
type dupFrame struct {
	seen map[ObjectKey]int
	kind dupContainerKind
}

// dupWalk is the mutable state of the duplicate-key token walk.
type dupWalk struct {
	stack     []dupFrame
	expectVal bool
}

// refuseDuplicateJSONKeys refuses the first key re-declared in the same
// enclosing object; each array element opens its own scope. The frame
// points at the refused file, pinned in docs/load.md, Refusals.
func refuseDuplicateJSONKeys(data SourceBytes, file FilePath) error {
	dec := jsontext.NewDecoder(bytes.NewReader(data), jsontext.AllowDuplicateNames(true))

	dw := &dupWalk{stack: nil, expectVal: false}

	for {
		tok, err := dec.ReadToken()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return jsonErr(data, file, err)
		}

		if err := dw.step(tok, lineAtByte(data, dec.InputOffset()), file); err != nil {
			return err
		}
	}
}

// step folds one token into the walk. A string is a key only when the
// innermost container is an object and no value is expected; inside an
// array every token is a value, so no string can forge an object key.
func (dw *dupWalk) step(tok jsontext.Token, line int, file FilePath) error {
	kind := tok.Kind()
	if kind == jsontext.KindBeginObject {
		dw.stack = append(dw.stack, dupFrame{kind: dupObject, seen: make(map[ObjectKey]int)})
		dw.expectVal = false

		return nil
	}

	if kind == jsontext.KindBeginArray {
		dw.stack = append(dw.stack, dupFrame{kind: dupArray, seen: nil})
		dw.expectVal = true

		return nil
	}

	if kind == jsontext.KindEndObject || kind == jsontext.KindEndArray {
		dw.pop()
		dw.expectVal = false

		return nil
	}

	if dw.expectVal {
		dw.expectVal = false

		return nil
	}

	return dw.stepKey(tok, line, file)
}

// pop closes the innermost container. JSON is well-formed at this point:
// the decoder has already refused mismatched delimiters.
func (dw *dupWalk) pop() {
	if len(dw.stack) > 0 {
		dw.stack = dw.stack[:len(dw.stack)-1]
	}
}

// stepKey records one object key, refusing re-declaration with the
// first-definition line named. Only object scopes have keys to record.
func (dw *dupWalk) stepKey(tok jsontext.Token, line int, file FilePath) error {
	if tok.Kind() != jsontext.KindString || len(dw.stack) == 0 {
		return nil
	}

	frame := &dw.stack[len(dw.stack)-1]
	if frame.kind != dupObject {
		return nil
	}

	key := tok.String()
	if first, dup := frame.seen[ObjectKey(key)]; dup {
		return refuseAt(ErrDuplicateKey, file, gorefuse.Line(first),
			gorefuse.Subject(`"`+key+`"`),
			gorefuse.Why(`delete one of the two "`+key+`" lines.`))
	}

	frame.seen[ObjectKey(key)] = line
	dw.expectVal = true

	return nil
}
