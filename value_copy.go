package goconf

import (
	"reflect"
	"strconv"

	"github.com/BurntSushi/toml"
)

type shadowContext struct {
	md       *toml.MetaData
	absorbed valueKeySet
}

// fillValueShadows copies the shadow decode into the real target and records
// every captured Value location for the undecoded-key filter.
func fillValueShadows(dst, shadow reflect.Value, md *toml.MetaData, src SourceBytes, file FilePath, key toml.Key) valueKeySet {
	ctx := &shadowContext{md: md, absorbed: valueKeySet{}}
	copyValueShadow(dst, shadow, key, src, file, ctx)

	return ctx.absorbed
}

type valueKeySet map[string]bool

// absorbs reports whether k sits inside one of the captured Value locations.
func (s valueKeySet) absorbs(k toml.Key) bool {
	for i := 1; i <= len(k); i++ {
		if s[k[:i].String()] {
			return true
		}
	}

	return false
}

type shadowCopier func(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext)

func copyValueShadow(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
	if dst.Type() == valueTypeOf() {
		copyIntoValue(dst, shadow, key, src, file, ctx)

		return
	}

	copiers := map[reflect.Kind]shadowCopier{
		reflect.Struct: copyStructShadow,
		reflect.Map:    copyMapShadow,
		reflect.Pointer: func(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
			copyIndirectShadow(dst, shadow, key, src, file, ctx, false)
		},
		reflect.Slice: func(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
			copyIndirectShadow(dst, shadow, key, src, file, ctx, true)
		},
		reflect.Array: func(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
			copyIndexedShadow(dst, shadow, ctx, key, src, file)
		},
	}

	if copier, ok := copiers[dst.Kind()]; ok {
		copier(dst, shadow, key, src, file, ctx)

		return
	}

	dst.Set(shadow)
}

func copyIntoValue(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
	prim, ok := reflect.TypeAssert[toml.Primitive](shadow)
	if !ok {
		return
	}

	target, ok := reflect.TypeAssert[*Value](dst.Addr())
	if !ok {
		return
	}

	target.setDecoded(prim, ctx.md, src, file, key)
	ctx.absorbed.mark(key)
}

func (s valueKeySet) mark(k toml.Key) {
	s[k.String()] = true
}

func copyStructShadow(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
	for i := range dst.NumField() {
		if !dst.Field(i).CanSet() {
			continue
		}

		seg := tagPathOf(fieldAt(dst.Type(), i))
		child := appendKey(key, metaSeg(seg))
		copyValueShadow(dst.Field(i), shadow.Field(i), child, src, file, ctx)
	}
}

func copyMapShadow(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext) {
	dst.Set(reflect.MakeMapWithSize(dst.Type(), shadow.Len()))

	for _, mk := range shadow.MapKeys() {
		child := appendKey(key, metaSeg(mk.String()))
		slot := reflect.New(dst.Type().Elem()).Elem()
		copyValueShadow(slot, shadow.MapIndex(mk), child, src, file, ctx)
		dst.SetMapIndex(mk.Convert(dst.Type().Key()), slot)
	}
}

func copyIndirectShadow(dst, shadow reflect.Value, key toml.Key, src SourceBytes, file FilePath, ctx *shadowContext, indexed bool) {
	if shadow.IsNil() {
		return
	}

	if indexed {
		dst.Set(reflect.MakeSlice(dst.Type(), shadow.Len(), shadow.Len()))
		copyIndexedShadow(dst, shadow, ctx, key, src, file)

		return
	}

	dst.Set(reflect.New(dst.Type().Elem()))
	copyValueShadow(dst.Elem(), shadow.Elem(), key, src, file, ctx)
}

func copyIndexedShadow(dst, shadow reflect.Value, ctx *shadowContext, key toml.Key, src SourceBytes, file FilePath) {
	for i := range shadow.Len() {
		child := appendKey(key, indexSegment(i))
		copyValueShadow(dst.Index(i), shadow.Index(i), child, src, file, ctx)
	}
}

func indexSegment(i int) metaSeg {
	return metaSeg("[" + strconv.Itoa(i) + "]")
}

func appendKey(key toml.Key, seg metaSeg) toml.Key {
	out := make(toml.Key, len(key), len(key)+1)
	copy(out, key)

	return append(out, string(seg))
}
