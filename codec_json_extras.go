package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"io"
	"reflect"

	"github.com/adambarber/gorefuse"
)

var (
	// errUnexpectedJSONToken indicates the token stream reached an unexpected token during schema-aware decode.
	ErrUnexpectedJSONToken = errors.New("unexpected token")
	// errExpectedJSONObjectKey indicates an object member key token was not a string.
	ErrExpectedJSONObjectKey = errors.New("expected object key")
	// errEmptyJSON indicates the JSON input contained no document content.
	errEmptyJSON = errors.New("empty JSON")
)

// decodeJSONStrictWithExtras decodes JSON into dst with schema-aware unknown
// field handling: keys not matching fixed fields are captured into an extras
// field when declared, otherwise refused.
func decodeJSONStrictWithExtras(data SourceBytes, file FilePath, dst Target) error {
	if err := jsonDecodePrecheck(data, file); err != nil {
		return err
	}

	dec := jsontext.NewDecoder(bytes.NewReader(data), jsontext.AllowDuplicateNames(true))

	v := derefValue(dst.Value)

	kind := dec.PeekKind()
	if kind == 0 {
		tok, err := dec.ReadToken()
		if err != nil {
			return jsonErr(data, file, err)
		}

		kind = tok.Kind()
	}

	if err := jsonExtrasRootShape(file, kind, v); err != nil {
		return err
	}

	if err := decodeJSONNextValue(dec, data, file, v, ""); err != nil {
		return err
	}

	if dec.PeekKind() != 0 {
		if _, err := dec.ReadToken(); err == nil {
			return refuseAt(ErrDecoderRefused, file, 0,
				gorefuse.Subject("data after the document root"),
				gorefuse.Why("one JSON document per file; remove the trailing bytes."))
		} else if !errors.Is(err, io.EOF) {
			return jsonErr(data, file, err)
		}
	}

	return nil
}

func decodeJSONNextValue(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	dst = derefAlloc(dst)

	kind := dec.PeekKind()
	if kind == 0 {
		vval, err := dec.ReadValue()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return jsonErr(data, file, err)
		}

		dec = jsontext.NewDecoder(bytes.NewReader(vval), jsontext.AllowDuplicateNames(true))
		kind = dec.PeekKind()
	}

	switch kind {
	case '{', '[':
		return decodeJSONContainerValue(dec, kind, data, file, dst, path)
	case '"', '0', 't', 'f', 'n':
		return decodeJSONScalar(dec, kind, data, file, dst, path)
	case jsontext.KindInvalid, '}', ']':
		return jsonUnexpectedToken(file)
	default:
		return jsonUnexpectedToken(file)
	}
}

func decodeJSONContainerValue(dec *jsontext.Decoder, kind jsontext.Kind, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	_, err := dec.ReadToken()
	if err != nil {
		return jsonErr(data, file, err)
	}

	if kind == '[' {
		return decodeJSONArray(dec, data, file, dst, path)
	}

	return decodeJSONObject(dec, data, file, dst, path)
}

func decodeJSONScalar(dec *jsontext.Decoder, kind jsontext.Kind, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	switch kind {
	case '"':
		tok, err := dec.ReadToken()
		if err != nil {
			return jsonErr(data, file, err)
		}

		return setJSONScalar(dst, jsonText(tok.String()), file, path)
	case '0':
		vval, err := dec.ReadValue()
		if err != nil {
			return jsonErr(data, file, err)
		}

		return setJSONNumberText(dst, jsonText(vval), file, path)
	case 't', 'f':
		tok, err := dec.ReadToken()
		if err != nil {
			return jsonErr(data, file, err)
		}

		return setJSONBool(dst, tok.Bool(), file, path)
	case 'n':
		if _, err := dec.ReadToken(); err != nil {
			return jsonErr(data, file, err)
		}

		dst.Set(reflect.Zero(dst.Type()))

		return nil
	case '{', '[', jsontext.KindInvalid, '}', ']':
		return jsonUnexpectedToken(file)
	}

	return jsonUnexpectedToken(file)
}
