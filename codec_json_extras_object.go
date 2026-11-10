// json extras object decoder: structs and extras capture.
package goconf

import (
	"encoding/json/jsontext"
	"reflect"

	"github.com/adambarber/gorefuse"
)

// decodeJSONObject routes extras-aware object decoding by destination shape.
// Struct targets may capture unknown members into an extras map field.

func decodeJSONObject(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
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

func decodeJSONObjectStruct(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	initExtrasMap(dst)

	for dec.PeekKind() != '}' {
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

func decodeJSONObjectAny(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	mv := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), emptyJSONValueType()))
	if err := decodeJSONIntoMap(dec, data, file, mv, path); err != nil {
		return err
	}

	dst.Set(mv)

	return nil
}

func decodeJSONObjectField(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath, key keySegment) error {
	sf, ok := matchJSONStructField(dst.Type(), key)
	if ok {
		child := childPath(path, tagPathOf(sf))
		return decodeJSONNextValue(dec, data, file, dst.FieldByIndex(sf.Index), child)
	}

	if ef, ok := extrasField(dst.Type()); ok {
		return decodeJSONExtra(dec, data, file, dst.FieldByIndex(ef.Index), path, key)
	}

	return jsonUnknownFieldRefusal(data, file, dec, key)
}

func decodeJSONExtra(dec *jsontext.Decoder, data SourceBytes, file FilePath, extras reflect.Value, path FieldPath, key keySegment) error {
	if extras.Kind() != reflect.Map {
		return jsonUnknownFieldRefusal(data, file, dec, key)
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	elem := reflect.New(extras.Type().Elem()).Elem()
	if err := decodeJSONNextValue(dec, data, file, elem, childPath(path, FieldPath(key))); err != nil {
		return err
	}

	extras.SetMapIndex(reflect.ValueOf(string(key)).Convert(extras.Type().Key()), elem)

	return nil
}

func jsonUnknownFieldRefusal(data SourceBytes, file FilePath, dec *jsontext.Decoder, key keySegment) error {
	line := lineAtByte(data, dec.InputOffset())

	return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
}

func consumeJSONObjectEnd(dec *jsontext.Decoder, data SourceBytes, file FilePath) error {
	if _, err := dec.ReadToken(); err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}

func emptyJSONValueType() reflect.Type {
	return reflect.TypeOf((*any)(nil)).Elem()
}

func decodeJSONIntoMap(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	if dst.IsNil() && dst.CanSet() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for dec.PeekKind() != '}' {
		if err := decodeJSONMapEntry(dec, data, file, dst, path); err != nil {
			return err
		}
	}

	_, err := dec.ReadToken()
	if err != nil {
		return jsonErr(data, file, err)
	}

	return nil
}

func decodeJSONMapEntry(dec *jsontext.Decoder, data SourceBytes, file FilePath, dst reflect.Value, path FieldPath) error {
	key, err := decodeJSONObjectKey(dec, data, file)
	if err != nil {
		return err
	}

	elem := reflect.New(dst.Type().Elem()).Elem()

	child := childPath(path, FieldPath(key))
	if err := decodeJSONNextValue(dec, data, file, elem, child); err != nil {
		return err
	}

	dst.SetMapIndex(reflect.ValueOf(key).Convert(dst.Type().Key()), elem)

	return nil
}
