package goconf

import "reflect"

func initExtrasMap(dst reflect.Value) {
	ef, ok := extrasField(dst.Type())
	if !ok {
		return
	}

	fv := dst.FieldByIndex(ef.Index)
	if fv.Kind() == reflect.Map && fv.IsNil() && fv.CanSet() {
		fv.Set(reflect.MakeMap(fv.Type()))
	}
}
