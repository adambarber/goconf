package goconf

import (
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

func tomlLookupMetaValue(raw tomlKeyMap, md *toml.MetaData, k toml.Key, aotCurrent, inlineCur map[tomlPathKey]int) reflect.Value {
	if len(k) == 0 {
		return reflect.Value{}
	}

	return tomlLookupFromPrimitives(reflect.ValueOf(raw), md, k, aotCurrent, inlineCur)
}

func tomlLookupFromPrimitives(v reflect.Value, md *toml.MetaData, k toml.Key, aotCurrent, inlineCur map[tomlPathKey]int) reflect.Value {
	v = derefReflect(v)
	if !v.IsValid() || len(k) == 0 {
		return v
	}

	if v.Kind() != reflect.Map {
		return tomlLookupDecodedInValue(v, k, aotCurrent, inlineCur, 0)
	}

	mv := tomlMapLookup(v, metaSeg(k[0]))
	if !mv.IsValid() {
		return reflect.Value{}
	}

	if len(k) == 1 {
		return tomlDecodePrimitiveValue(md, mv)
	}

	decoded := tomlDecodePrimitiveValue(md, mv)
	if !decoded.IsValid() {
		return reflect.Value{}
	}

	return tomlLookupDecodedInValue(decoded, k[1:], aotCurrent, inlineCur, 1)
}

func tomlLookupDecodedInValue(v reflect.Value, k toml.Key, aotCurrent, inlineCur map[tomlPathKey]int, offset int) reflect.Value {
	for j, seg := range k {
		i := offset + j

		v = derefReflect(v)
		if !v.IsValid() {
			return reflect.Value{}
		}

		next, ok := tomlStepDecodedValue(v, k, i, metaSeg(seg), aotCurrent, inlineCur)
		if !ok {
			return reflect.Value{}
		}

		v = next
	}

	return derefReflect(v)
}

func tomlStepDecodedValue(v reflect.Value, k toml.Key, i int, seg metaSeg, aotCurrent, inlineCur map[tomlPathKey]int) (reflect.Value, bool) {
	if v.Kind() == reflect.Struct {
		fv, ok := structFieldByPathName(v, seg)

		return fv, ok
	}

	if v.Kind() == reflect.Map {
		return tomlMapLookup(v, seg), true
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return tomlStepSliceDecoded(v, k, i, aotCurrent, inlineCur), true
	}

	return reflect.Value{}, false
}

func tomlStepSliceDecoded(v reflect.Value, k toml.Key, i int, aotCurrent, inlineCur map[tomlPathKey]int) reflect.Value {
	if v.Len() == 0 {
		return reflect.Value{}
	}

	pk := tomlPathKey(strings.Join(k[:i+1], "."))
	idx := v.Len() - 1

	if aotIdx, ok := aotCurrent[pk]; ok && aotIdx >= 0 && aotIdx < v.Len() {
		idx = aotIdx
	} else if inIdx, ok := inlineCur[pk]; ok && inIdx >= 0 && inIdx < v.Len() {
		idx = inIdx
	}

	return v.Index(idx)
}
