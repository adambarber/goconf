package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"strings"

	"github.com/adambarber/gorefuse"
)

// lineAtByte returns the 1-based line number of the byte at offset in data.
// JSON errors report byte offsets; refusals name lines, so the library
// translates at the boundary instead of letting offsets leak to callers.
func lineAtByte(data SourceBytes, offset int64) int {
	if offset < 0 {
		return 1
	}

	if offset > int64(len(data)) {
		offset = int64(len(data))
	}

	return bytes.Count(data[:offset], []byte("\n")) + 1
}

// jsonErr converts a decode error into a refusal naming the line the
// decoder stopped on, using the raw bytes it was reading. An unknown-field
// refusal names the key's line, found by re-walking the token stream.
func jsonErr(data SourceBytes, file FilePath, err error) error {
	var syntax *jsontext.SyntacticError

	var sem *json.SemanticError

	switch {
	case errors.As(err, &syntax):
		return jsonSyntacticErr(data, file, syntax)
	case errors.As(err, &sem):
		return jsonSemanticErr(data, file, sem)
	default:
		if len(data) == 0 {
			return refuseDecoder(
				gorefuse.At(gorefuse.File(file), 0,
					gorefuse.Subject("JSON syntax"),
					gorefuse.Why("the file is empty; add settings or delete it.")), err)
		}

		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON syntax"),
				syntaxCheck()), err)
	}
}

func jsonSyntacticErr(data SourceBytes, file FilePath, syntax *jsontext.SyntacticError) error {
	line := lineAtByte(data, syntax.ByteOffset)
	msg := syntax.Error()

	if strings.Contains(msg, "after top-level value") {
		return refuseAt(ErrDecoderRefused, file, 0,
			gorefuse.Subject("data after the document root"),
			gorefuse.Why("one JSON document per file; remove the trailing bytes."))
	}

	if strings.Contains(msg, "duplicate object member name") {
		key := jsonPointerLastKey(syntax.JSONPointer)
		if key != "" {
			first := jsonKeyLine(data, key)
			if first == 0 {
				first = line
			}

			return refuseAt(ErrDuplicateKey, file, gorefuse.Line(first),
				gorefuse.Subject(`"`+string(key)+`"`),
				gorefuse.Why(`delete one of the two "`+string(key)+`" lines.`))
		}
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON syntax"),
				gorefuse.Why("the file is empty; add settings or delete it.")), syntax)
	}

	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), gorefuse.Line(line),
			gorefuse.Subject("JSON syntax"),
			syntaxCheck()), syntax)
}

func jsonSemanticErr(data SourceBytes, file FilePath, sem *json.SemanticError) error {
	line := lineAtByte(data, sem.ByteOffset)

	msg := sem.Error()
	if strings.Contains(msg, "unknown object member name") {
		key := jsonPointerLastKey(sem.JSONPointer)
		if key == "" {
			key = jsonErrQuotedName(refusalText(msg))
		}

		if key != "" {
			if found := jsonKeyLine(data, key); found > 0 {
				line = found
			}

			return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
				gorefuse.Subject(`"`+key+`"`),
				gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
		}
	}

	return jsonSemanticValueErr(file, sem, line)
}

func jsonSemanticValueErr(file FilePath, sem *json.SemanticError, line int) error {
	key := jsonPointerLastKey(sem.JSONPointer)
	if sem.JSONPointer == "" || key == "" {
		if isJSONRootKind(sem.JSONKind) {
			return jsonRootShapeRefusal(file, line, jsonKindFoundName(sem.JSONKind))
		}
	}

	if sem.GoType != nil {
		found := jsonKindFoundName(sem.JSONKind)
		if found == "" {
			found = goTypeName("value")
		}

		if sem.JSONPointer == "" {
			return jsonRootShapeRefusal(file, line, found)
		}

		return typeMismatchRefusal(file, line, key,
			jsonWantKind(sem.GoType), jsonFoundTerm(found))
	}

	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), gorefuse.Line(line),
			gorefuse.Subject("JSON syntax"),
			syntaxCheck()), sem)
}

/* kind and pointer helpers moved to codec_json_err_helpers.go */

/* syntax and term helpers moved to codec_json_err_terms.go */

func jsonUnexpectedToken(file FilePath) error {
	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), 0, gorefuse.Subject("JSON syntax"), syntaxCheck()),
		ErrUnexpectedJSONToken,
	)
}
