package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/adambarber/gorefuse"
)

var (
	// errUnexpectedJSONToken indicates the token stream reached an unexpected token during schema-aware decode.
	ErrUnexpectedJSONToken = errors.New("unexpected token")
	// errExpectedJSONObjectKey indicates an object member key token was not a string.
	ErrExpectedJSONObjectKey = errors.New("expected object key")
)

// decodeJSONStrictWithExtras decodes JSON into dst with schema-aware unknown
// field handling: keys not matching fixed fields are captured into an extras
// field when declared, otherwise refused.
func decodeJSONStrictWithExtras(data SourceBytes, file FilePath, dst Target) error {
	if err := refuseDuplicateJSONKeys(data, file); err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		if len(data) == 0 {
			return refuseDecoder(
				gorefuse.At(gorefuse.File(file), 0,
					gorefuse.Subject("JSON syntax"),
					gorefuse.Why("the file is empty; add settings or delete it.")), err)
		}

		return jsonErr(data, file, err)
	}

	v := derefValue(dst.Value)

	if err := jsonExtrasRootShape(file, tok, v); err != nil {
		return err
	}

	if err := decodeJSONValue(dec, tok, data, file, v, ""); err != nil {
		return err
	}

	_, tailErr := dec.Token()
	if tailErr == nil {
		return refuseAt(ErrDecoderRefused, file, 0,
			gorefuse.Subject("data after the document root"),
			gorefuse.Why("one JSON document per file; remove the trailing bytes."))
	}

	return nil
}

func decodeJSONValue(dec *json.Decoder, tok json.Token, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	dst = derefAlloc(dst)

	switch d := tok.(type) {
	case json.Delim:
		return decodeJSONDelim(dec, d, data, file, dst, path)
	case string:
		return setJSONScalar(dst, jsonText(d), file, path)
	case json.Number:
		return setJSONNumber(dst, d, file, path)
	case bool:
		return setJSONBool(dst, d, file, path)
	case nil:
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	default:
		return refuseDecoder(gorefuse.At(gorefuse.File(file), 0, gorefuse.Subject("JSON syntax"), syntaxCheck()), ErrUnexpectedJSONToken)
	}
}

func decodeJSONDelim(dec *json.Decoder, d json.Delim, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	switch d {
	case '{':
		return decodeJSONObject(dec, data, file, dst, path)
	case '[':
		return decodeJSONArray(dec, data, file, dst, path)
	default:
		return refuseDecoder(gorefuse.At(gorefuse.File(file), 0, gorefuse.Subject("JSON syntax"), syntaxCheck()), ErrUnexpectedJSONToken)
	}
}

func decodeJSONObject(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	if dst.Kind() == reflect.Map {
		return decodeJSONIntoMap(dec, data, file, dst, path)
	}

	if dst.Kind() == reflect.Interface && dst.Type().NumMethod() == 0 {
		return decodeJSONObjectAny(dec, data, file, dst, path)
	}

	if dst.Kind() == reflect.Struct {
		return decodeJSONObjectStruct(dec, data, file, dst, path)
	}

	return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindTable))
}

func decodeJSONObjectStruct(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	initExtrasMap(dst)

	for dec.More() {
		key, keyErr := decodeJSONObjectKey(dec, data, file)
		if keyErr != nil {
			return keyErr
		}

		fieldErr := decodeJSONObjectField(dec, data, file, dst, path, key)
		if fieldErr != nil {
			return fieldErr
		}
	}

	return consumeJSONObjectEnd(dec, data, file)
}

func decodeJSONObjectAny(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	mv := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), emptyJSONValueType()))
	if err := decodeJSONIntoMap(dec, data, file, mv, path); err != nil {
		return err
	}

	dst.Set(mv)

	return nil
}

func decodeJSONObjectField(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath, key keySegment) error {
	// peek/read value token by decoding raw? easier to read token for scalars, but values can be objects/arrays.
	// Read next token to dispatch.
	vtok, err := dec.Token()
	if err != nil {
		return jsonErr(data, file, err)
	}

	sf, ok := matchStructField(dst.Type(), key, jsonFieldMatches)
	if ok {
		child := childPath(path, tagPathOf(sf))
		return decodeJSONValue(dec, vtok, data, file, dst.FieldByIndex(sf.Index), child)
	}

	if ef, ok := extrasField(dst.Type()); ok {
		return decodeJSONExtra(dec, vtok, data, file, dst.FieldByIndex(ef.Index), path, key)
	}

	return jsonUnknownFieldRefusal(data, file, dec, key)
}

func decodeJSONExtra(dec *json.Decoder, vtok json.Token, data SourceBytes, file FilePath, extras reflect.Value, path FieldPath, key keySegment) error {
	if extras.Kind() != reflect.Map {
		return jsonUnknownFieldRefusal(data, file, dec, key)
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	elem := reflect.New(extras.Type().Elem()).Elem()
	if err := decodeJSONValue(dec, vtok, data, file, elem, childPath(path, FieldPath(key))); err != nil {
		return err
	}

	extras.SetMapIndex(reflect.ValueOf(string(key)).Convert(extras.Type().Key()), elem)

	return nil
}

func jsonUnknownFieldRefusal(data SourceBytes, file FilePath, dec *json.Decoder, key keySegment) error {
	line := lineAtByte(data, dec.InputOffset())

	return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
}

func consumeJSONObjectEnd(dec *json.Decoder, data SourceBytes, file FilePath) error {
	if _, err := dec.Token(); err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}
