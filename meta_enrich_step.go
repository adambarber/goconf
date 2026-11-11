package goconf

import "reflect"

func stepPathValue(v reflect.Value, seg pathSeg) (reflect.Value, bool) {
	v = derefReflect(v)
	if !v.IsValid() {
		return reflect.Value{}, false
	}

	if seg.isIdx {
		return stepIndexValue(v, seg.index)
	}

	return stepNamedValue(v, seg.name)
}

func stepIndexValue(v reflect.Value, index int) (reflect.Value, bool) {
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return reflect.Value{}, false
	}

	if index < 0 || index >= v.Len() {
		return reflect.Value{}, false
	}

	return derefReflect(v.Index(index)), true
}

func stepNamedValue(v reflect.Value, name metaSeg) (reflect.Value, bool) {
	if v.Kind() == reflect.Struct {
		return structFieldByPathName(v, name)
	}

	if v.Kind() == reflect.Map {
		return mapValueByPathName(v, name)
	}

	return reflect.Value{}, false
}

func derefReflect(v reflect.Value) reflect.Value {
	for v.IsValid() {
		k := v.Kind()
		if k != reflect.Pointer && k != reflect.Interface {
			return v
		}

		if v.IsNil() {
			return reflect.Value{}
		}

		v = v.Elem()
	}

	return v
}

func structFieldByPathName(v reflect.Value, name metaSeg) (reflect.Value, bool) {
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		if metaSeg(tagPathOf(&f)) == name {
			return derefReflect(v.Field(i)), true
		}
	}

	return reflect.Value{}, false
}

func mapValueByPathName(v reflect.Value, name metaSeg) (reflect.Value, bool) {
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return reflect.Value{}, false
	}

	key := reflect.ValueOf(string(name))
	if key.Type().AssignableTo(v.Type().Key()) {
		return mapIndexValue(v, key)
	}

	if key.Type().ConvertibleTo(v.Type().Key()) {
		return mapIndexValue(v, key.Convert(v.Type().Key()))
	}

	return reflect.Value{}, false
}

func mapIndexValue(v, key reflect.Value) (reflect.Value, bool) {
	mv := v.MapIndex(key)
	if !mv.IsValid() {
		return reflect.Value{}, false
	}

	return derefReflect(mv), true
}
