package goconf

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/adambarber/gorefuse"
)

// Rendered is one safe rendering of a resolved config.
type Rendered string

// renderValue is Redact's reflection body; it keeps the generic boundary in
// load.go so the bare type-parameter constraint stays in its sanctioned file.
func renderValue(v reflect.Value) Rendered {
	var b strings.Builder

	writeRendered(&b, v, secretUnmarked)

	return Rendered(b.String())
}

func writeRendered(b *strings.Builder, v reflect.Value, decision secretDecision) {
	if !v.IsValid() {
		b.WriteString("<nil>")
		return
	}

	if decision == secretMarked {
		b.WriteString(gorefuse.Redacted)
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		writeIndirect(b, v, decision)
	case reflect.Struct:
		writeStruct(b, v, decision)
	case reflect.Slice, reflect.Array:
		writeList(b, v, decision)
	case reflect.Map:
		writeMap(b, v, decision)
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.String, reflect.UnsafePointer:
		_, _ = fmt.Fprintf(b, "%+v", v.Interface())
	}
}

func writeIndirect(b *strings.Builder, v reflect.Value, decision secretDecision) {
	if v.IsNil() {
		b.WriteString("<nil>")
		return
	}

	if v.Kind() == reflect.Pointer {
		b.WriteByte('&')
	}

	writeRendered(b, v.Elem(), decision)
}

func writeStruct(b *strings.Builder, v reflect.Value, decision secretDecision) {
	b.WriteByte('{')

	t := v.Type()
	for i := range t.NumField() {
		if i > 0 {
			b.WriteByte(' ')
		}

		f := t.Field(i)
		b.WriteString(f.Name)
		b.WriteByte(':')

		if !f.IsExported() {
			_, _ = fmt.Fprintf(b, "%+v", v.Field(i))
			continue
		}

		next := applySecretTag(declaredSecret(&f), decision)
		writeRendered(b, v.Field(i), next)
	}

	b.WriteByte('}')
}

func writeList(b *strings.Builder, v reflect.Value, decision secretDecision) {
	b.WriteByte('[')

	for i := range v.Len() {
		if i > 0 {
			b.WriteByte(' ')
		}

		writeRendered(b, v.Index(i), decision)
	}

	b.WriteByte(']')
}

func writeMap(b *strings.Builder, v reflect.Value, decision secretDecision) {
	b.WriteString("map[")

	for i, key := range sortedKeys(v) {
		if i > 0 {
			b.WriteByte(' ')
		}

		writeRendered(b, key, secretUnmarked)
		b.WriteByte(':')
		writeRendered(b, v.MapIndex(key), decision)
	}

	b.WriteByte(']')
}

func sortedKeys(v reflect.Value) []reflect.Value {
	keys := make([]reflect.Value, 0, v.Len())

	iter := v.MapRange()
	for iter.Next() {
		keys = append(keys, iter.Key())
	}

	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})

	return keys
}
