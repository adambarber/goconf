package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
)

func tomlValueConvertible(val reflect.Value, elemType reflect.Type) bool {
	if !val.Type().ConvertibleTo(elemType) {
		return false
	}

	if val.Kind() == elemType.Kind() {
		return true
	}

	if tomlIntFamily(val.Kind()) && tomlIntFamily(elemType.Kind()) {
		return tomlIntFits(val, elemType)
	}

	if tomlUintFamily(val.Kind()) && tomlUintFamily(elemType.Kind()) {
		return tomlUintFits(val, elemType)
	}

	if tomlFloatKind(val.Kind()) && tomlFloatKind(elemType.Kind()) {
		return tomlFloatFits(val, elemType)
	}

	return false
}

func tomlIntFamily(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func tomlUintFamily(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

func tomlFloatKind(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

func tomlIntFits(val reflect.Value, elemType reflect.Type) bool {
	tmp := reflect.New(elemType).Elem()
	if tmp.OverflowInt(val.Int()) {
		return false
	}

	tmp.SetInt(val.Int())

	return tmp.Int() == val.Int()
}

func tomlUintFits(val reflect.Value, elemType reflect.Type) bool {
	tmp := reflect.New(elemType).Elem()

	return !tmp.OverflowUint(val.Uint())
}

func tomlFloatFits(val reflect.Value, elemType reflect.Type) bool {
	tmp := reflect.New(elemType).Elem()

	return !tmp.OverflowFloat(val.Float())
}

func tomlRawValueTerm(v reflect.Value) userTerm {
	if !v.IsValid() {
		return ""
	}

	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map, reflect.Struct:
		return kindUserTerm(kindTable)
	case reflect.Slice, reflect.Array:
		return kindUserTerm(kindList)
	case reflect.Bool:
		return kindUserTerm(kindBool)
	case reflect.String:
		return kindUserTerm(kindString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return kindUserTerm(kindFloat)
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Pointer, reflect.UnsafePointer:
		return ""
	}

	return ""
}

func decodeTOMLRaw(data SourceBytes, file FilePath) (reflect.Value, error) {
	rv := reflect.New(reflect.TypeOf((*interface{})(nil)).Elem())
	if _, err := toml.Decode(string(data), rv.Interface()); err != nil {
		return reflect.Value{}, tomlDecodeRefusal(file, err)
	}

	return rv.Elem(), nil
}
