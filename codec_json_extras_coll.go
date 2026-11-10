// extras-aware JSON array decoding helpers.
package goconf

import (
	"encoding/json/jsontext"
	"reflect"
)

// decodeJSONArray decodes a JSON array token stream into dst for extras-aware paths.

func decodeJSONArray(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	if dst.Kind() == reflect.Interface && dst.Type().NumMethod() == 0 {
		return decodeJSONArrayAny(dec, data, file, dst, path)
	}

	if dst.Kind() != reflect.Slice && dst.Kind() != reflect.Array {
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindList))
	}

	return decodeJSONArrayTyped(dec, data, file, dst, path)
}

func decodeJSONArrayAny(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	sv := reflect.MakeSlice(reflect.SliceOf(emptyJSONValueType()), 0, 0)

	for dec.PeekKind() != ']' {
		elem := reflect.New(emptyJSONValueType()).Elem()
		if err := decodeJSONNextValue(dec, data, file, elem, path); err != nil {
			return err
		}

		sv = reflect.Append(sv, elem)
	}

	if _, err := dec.ReadToken(); err != nil {
		return jsonErr(data, file, err)
	}

	dst.Set(sv)

	return nil
}

func decodeJSONArrayTyped(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	idx := 0

	for dec.PeekKind() != ']' {
		var err error

		dst, err = decodeJSONArrayElem(dec, data, file, dst, path, idx)
		if err != nil {
			return err
		}

		idx++
	}

	if _, err := dec.ReadToken(); err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}

func decodeJSONArrayElem(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath, idx int) (reflect.Value, error) {
	if dst.Kind() == reflect.Slice {
		if idx >= dst.Len() {
			dst.Set(reflect.Append(dst, reflect.Zero(dst.Type().Elem())))
		}
	} else if idx >= dst.Len() {
		elem := reflect.New(dst.Type().Elem()).Elem()
		if err := decodeJSONNextValue(dec, data, file, elem, path); err != nil {
			return dst, err
		}

		return dst, jsonFixedArrayOverflow(file, path)
	}

	if err := decodeJSONNextValue(dec, data, file, dst.Index(idx), path); err != nil {
		return dst, err
	}

	return dst, nil
}

/* jsonFixedArrayOverflow moved to codec_json_extras_scalar.go */
