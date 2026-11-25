package goconf

import "reflect"

// cloneValue deep-copies a default so one Load cannot mutate the next.
func cloneValue(v reflect.Value) reflect.Value {
	switch {
	case v.Kind() == reflect.Pointer:
		return clonePointer(v)
	case v.Kind() == reflect.Map:
		return cloneMap(v)
	case v.Kind() == reflect.Slice:
		return cloneSlice(v)
	case v.Kind() == reflect.Interface:
		return cloneInterface(v)
	case v.Kind() == reflect.Struct:
		return cloneStruct(v)
	default:
		return v
	}
}

// clonePointer copies through a pointer so the target never shares it.
func clonePointer(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	copied := reflect.New(v.Type().Elem())
	copied.Elem().Set(cloneValue(v.Elem()))

	return copied
}

// cloneInterface copies the captured value without altering its interface type.
func cloneInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	copied := reflect.New(v.Type()).Elem()
	copied.Set(cloneValue(v.Elem()))

	return copied
}

// cloneMap copies each key and value so the result shares no backing store.
func cloneMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	out := reflect.MakeMapWithSize(v.Type(), v.Len())

	iter := v.MapRange()
	for iter.Next() {
		out.SetMapIndex(cloneValue(iter.Key()), cloneValue(iter.Value()))
	}

	return out
}

// cloneSlice copies each element into fresh storage.
func cloneSlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}

	out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	for i := range v.Len() {
		out.Index(i).Set(cloneValue(v.Index(i)))
	}

	return out
}

// cloneStruct copies schema fields one by one; an opaque leaf is an immutable
// value whose whole-struct copy is enough.
func cloneStruct(v reflect.Value) reflect.Value {
	if !isSchemaStruct(v) {
		return v
	}

	out := reflect.New(v.Type()).Elem()

	t := v.Type()
	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !isDefaultableField(f) || !v.Field(i).CanInterface() {
			continue
		}

		out.Field(i).Set(cloneValue(v.Field(i)))
	}

	return out
}
