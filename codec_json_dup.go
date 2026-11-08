package goconf

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/adambarber/gorefuse"
)

// ObjectKey is one JSON object key exactly as it appears in the document.
type ObjectKey string

// dupScope is one open JSON object: the keys already declared in it,
// mapped to the line of first definition.
type dupScope struct {
	seen map[ObjectKey]int
}

// dupWalk is the mutable state of the duplicate-key token walk.
type dupWalk struct {
	scopes    []dupScope
	expectVal bool
}

// refuseDuplicateJSONKeys refuses the first key re-declared in the same
// enclosing object; each array element opens its own scope. The frame
// points at the refused file, pinned in docs/load.md, Refusals.
func refuseDuplicateJSONKeys(data SourceBytes, file FilePath) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	dw := &dupWalk{scopes: nil, expectVal: false}

	for {
		tok, err := dec.Token()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return jsonErr(data, file, err)
		}

		if err := dw.step(tok, lineAtByte(data, dec.InputOffset()), file); err != nil {
			return err
		}
	}
}

// step folds one token into the duplicate-key walk: delimiters push and
// pop object scopes, keys in the top scope are refused on re-declaration.
func (dw *dupWalk) step(tok json.Token, line int, file FilePath) error {
	if d, isDelim := tok.(json.Delim); isDelim {
		return dw.stepDelim(d)
	}

	if dw.expectVal {
		dw.expectVal = false

		return nil
	}

	return dw.stepKey(tok, line, file)
}

// stepDelim handles one structural token: { pushes a scope, } pops one;
// brackets carry no scope — array elements open their own object scopes
// and a close bracket returns to the enclosing scope untouched.
func (dw *dupWalk) stepDelim(d json.Delim) error {
	switch d {
	case '{':
		dw.scopes = append(dw.scopes, dupScope{seen: make(map[ObjectKey]int)})
	case '}':
		if len(dw.scopes) > 0 {
			dw.scopes = dw.scopes[:len(dw.scopes)-1]
		}
	}

	dw.expectVal = false

	return nil
}

// stepKey records one object key in the top scope, refusing the key when
// it was already declared there, with the first-definition line named.
func (dw *dupWalk) stepKey(tok json.Token, line int, file FilePath) error {
	key, isKey := tok.(string)
	if !isKey || len(dw.scopes) == 0 {
		return nil
	}

	scope := &dw.scopes[len(dw.scopes)-1]
	if first, dup := scope.seen[ObjectKey(key)]; dup {
		return refuseAt(ErrDuplicateKey, file, gorefuse.Line(first),
			gorefuse.Subject(`"`+key+`"`),
			gorefuse.Why(`delete one of the two "`+key+`" lines.`))
	}

	scope.seen[ObjectKey(key)] = line
	dw.expectVal = true

	return nil
}
