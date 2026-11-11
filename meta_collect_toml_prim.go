package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
)

func tomlDecodePrimitiveValue(md *toml.MetaData, mv reflect.Value) reflect.Value {
	if !mv.IsValid() {
		return reflect.Value{}
	}

	prim, ok := reflect.TypeAssert[toml.Primitive](mv)
	if !ok {
		return derefReflect(mv)
	}

	for _, dec := range tomlPrimDecoders() {
		if v, ok := dec(md, prim); ok {
			return v
		}
	}

	return reflect.Value{}
}

type tomlPrimDecoder func(*toml.MetaData, toml.Primitive) (reflect.Value, bool)

type tomlScalar interface {
	~int64 | ~float64 | ~bool
}

func tomlPrimDecoders() []tomlPrimDecoder {
	return []tomlPrimDecoder{
		tomlDecodePrimAsString,
		tomlDecodePrimScalar[int64],
		tomlDecodePrimScalar[float64],
		tomlDecodePrimScalar[bool],
		tomlDecodePrimAsStringSlice,
		tomlDecodePrimAsIntSlice,
	}
}

func tomlDecodePrimScalar[T tomlScalar](md *toml.MetaData, prim toml.Primitive) (reflect.Value, bool) {
	var v T
	if err := md.PrimitiveDecode(prim, &v); err != nil {
		return reflect.Value{}, false
	}

	return reflect.ValueOf(v), true
}

func tomlDecodePrimAsString(md *toml.MetaData, prim toml.Primitive) (reflect.Value, bool) {
	var s metaSeg
	if err := md.PrimitiveDecode(prim, &s); err != nil {
		return reflect.Value{}, false
	}

	return reflect.ValueOf(string(s)), true
}

func tomlDecodePrimAsStringSlice(md *toml.MetaData, prim toml.Primitive) (reflect.Value, bool) {
	var v []metaSeg
	if err := md.PrimitiveDecode(prim, &v); err != nil {
		return reflect.Value{}, false
	}

	out := append([]metaSeg(nil), v...)

	return reflect.ValueOf(out), true
}

func tomlDecodePrimAsIntSlice(md *toml.MetaData, prim toml.Primitive) (reflect.Value, bool) {
	ints := make([]int64, 0)
	if err := md.PrimitiveDecode(prim, &ints); err != nil {
		return reflect.Value{}, false
	}

	return reflect.ValueOf(ints), true
}

func tomlMapLookup(v reflect.Value, seg metaSeg) reflect.Value {
	if v.Kind() != reflect.Map {
		return reflect.Value{}
	}

	key := reflect.ValueOf(string(seg))
	if key.Type().AssignableTo(v.Type().Key()) {
		return v.MapIndex(key)
	}

	if key.Type().ConvertibleTo(v.Type().Key()) {
		return v.MapIndex(key.Convert(v.Type().Key()))
	}

	iter := v.MapRange()
	for iter.Next() {
		if mk := iter.Key(); mk.Kind() == reflect.String && mk.String() == string(seg) {
			return iter.Value()
		}
	}

	return reflect.Value{}
}
