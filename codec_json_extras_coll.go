package goconf

import (
	"encoding/json"
	"reflect"
)

// decodeJSONArray decodes a JSON array token stream into dst for extras-aware paths.

func decodeJSONArray(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	if dst.Kind() == reflect.Interface && dst.Type().NumMethod() == 0 {
		return decodeJSONArrayAny(dec, data, file, dst, path)
	}

	if dst.Kind() != reflect.Slice && dst.Kind() != reflect.Array {
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindList))
	}

	return decodeJSONArrayTyped(dec, data, file, dst, path)
}

func decodeJSONArrayAny(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	sv := reflect.MakeSlice(reflect.SliceOf(emptyJSONValueType()), 0, 0)

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return jsonErr(data, file, err)
		}

		elem := reflect.New(emptyJSONValueType()).Elem()
		if err := decodeJSONValue(dec, tok, data, file, elem, path); err != nil {
			return err
		}

		sv = reflect.Append(sv, elem)
	}

	if _, err := dec.Token(); err != nil {
		return jsonErr(data, file, err)
	}

	dst.Set(sv)

	return nil
}

func decodeJSONArrayTyped(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	idx := 0

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return jsonErr(data, file, err)
		}

		dst, err = decodeJSONArrayElem(dec, tok, data, file, dst, path, idx)
		if err != nil {
			return err
		}

		idx++
	}

	if _, err := dec.Token(); err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}

func decodeJSONArrayElem(dec *json.Decoder, tok json.Token, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath, idx int) (reflect.Value, error) {
	if idx >= dst.Len() && dst.Kind() == reflect.Slice {
		dst.Set(reflect.Append(dst, reflect.Zero(dst.Type().Elem())))
	}

	if idx < dst.Len() {
		if err := decodeJSONValue(dec, tok, data, file, dst.Index(idx), path); err != nil {
			return dst, err
		}
	}

	return dst, nil
}

func emptyJSONValueType() reflect.Type {
	return reflect.TypeOf((*any)(nil)).Elem()
}

func decodeJSONIntoMap(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	if dst.IsNil() && dst.CanSet() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for dec.More() {
		if err := decodeJSONMapEntry(dec, data, file, dst, path); err != nil {
			return err
		}
	}

	_, err := dec.Token()
	if err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}

func decodeJSONMapEntry(dec *json.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	key, err := decodeJSONObjectKey(dec, data, file)
	if err != nil {
		return err
	}

	vtok, err := dec.Token()
	if err != nil {
		return jsonErr(data, file, err)
	}

	elem := reflect.New(dst.Type().Elem()).Elem()

	child := childPath(path, FieldPath(key))
	if err := decodeJSONValue(dec, vtok, data, file, elem, child); err != nil {
		return err
	}

	dst.SetMapIndex(reflect.ValueOf(key).Convert(dst.Type().Key()), elem)

	return nil
}
