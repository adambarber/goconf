package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

// Value is an opaque TOML value that may hold a scalar or a table. It is
// meaningful only in TOML targets; JSON and YAML refuse fields of this type.
// Use DecodeValue to materialize it into a concrete type.
type Value struct {
	primitive toml.Primitive
	md        *toml.MetaData
	src       SourceBytes
	file      FilePath
	key       toml.Key
}

// setDecoded attaches the decode context goconf captures after the TOML
// parse; it is the only way a Value's private fields are populated.
func (v *Value) setDecoded(p toml.Primitive, md *toml.MetaData, src SourceBytes, file FilePath, key toml.Key) {
	v.primitive = p
	v.md = md
	v.src = src
	v.file = file
	v.key = key
}

// DecodeValue materializes one Value into T. Unknown keys inside a held table
// are refused with ErrUnknownField; the scalar and table shapes are the only
// ones this bounded container supports.
func DecodeValue[T any](v Value) (T, error) {
	return decodeValue[T](&v)
}

func decodeValue[T any](v *Value) (T, error) {
	var zero T

	if v.md == nil || reflect.ValueOf(v.primitive).IsZero() {
		return zero, refuseWhy(ErrDecoderRefused,
			gorefuse.Subject("an empty value container"),
			gorefuse.Why("fill the goconf.Value from a TOML source before decoding it."))
	}

	var table valueTable
	if err := v.md.PrimitiveDecode(v.primitive, &table); err == nil && table != nil {
		return decodeValueTable[T](v, table)
	}

	return decodeValueScalar[T](v)
}

// decodeValueScalar decodes the held primitive straight into the requested
// destination, so TOML integer, float, boolean, and string scalars all work.
func decodeValueScalar[T any](v *Value) (T, error) {
	var zero T

	dst := newValueDestination[T]()
	if err := v.md.PrimitiveDecode(v.primitive, dst.Addr().Interface()); err != nil {
		return zero, valueTypeRefusal(v, "the value's type")
	}

	return castValueDestination[T](dst)
}

// valueName is one key inside a held table; valueTable is that table's shape.
type valueName string

type valueTable map[valueName]toml.Primitive

// valueWant names the shape a destination expects, in user terms.
type valueWant string

func requireValueShape[T any](v *Value, kind reflect.Kind, want valueWant) (reflect.Value, error) {
	dst := newValueDestination[T]()
	if dst.Kind() != kind {
		return reflect.Value{}, valueTypeRefusal(v, want)
	}

	return dst, nil
}

func decodeValueTable[T any](v *Value, table valueTable) (T, error) {
	var zero T

	dst, err := requireValueShape[T](v, reflect.Struct, "a table")
	if err != nil {
		return zero, err
	}

	for name, prim := range table {
		f, ok := fieldMatching(dst.Type(), keySegment(string(name)))
		if !ok {
			return zero, refuseAt(ErrUnknownField, v.file, valueLine(v, name),
				gorefuse.Subject(`"`+name+`"`),
				gorefuse.Why("remove it, or declare it in the destination's schema."))
		}

		field := dst.FieldByIndex(f.Index)
		if err := v.md.PrimitiveDecode(prim, field.Addr().Interface()); err != nil {
			return zero, valueTypeRefusal(v, "the field's type")
		}
	}

	return castValueDestination[T](dst)
}

// newValueDestination allocates a T, dereferencing one pointer level so a
// pointer destination materializes into its element.
func newValueDestination[T any]() reflect.Value {
	dst := reflect.New(reflect.TypeFor[T]()).Elem()
	if dst.Kind() == reflect.Pointer {
		return reflect.New(dst.Type().Elem()).Elem()
	}

	return dst
}

func castValueDestination[T any](dst reflect.Value) (T, error) {
	got, ok := reflect.TypeAssert[T](dst)
	if !ok {
		var zero T

		return zero, refuseWhy(ErrTargetCast,
			gorefuse.Subject("the value destination"),
			gorefuse.Why("the destination does not hold the requested type; this is a construction fault."))
	}

	return got, nil
}

func valueTypeRefusal(v *Value, want valueWant) error {
	return refuseAt(ErrTypeMismatch, v.file, valueLine(v, ""),
		gorefuse.Subject("a value that cannot fill "+string(want)),
		gorefuse.Why("give the destination the value's type."))
}

func valueLine(v *Value, child valueName) gorefuse.Line {
	key := v.key
	if child != "" {
		key = append(append(toml.Key{}, key...), string(child))
	}

	line, _ := tomlFindKeyLine(v.src, 0, key, v.md)

	return gorefuse.Line(line)
}
