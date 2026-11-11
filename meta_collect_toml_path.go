package goconf

import (
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

func canonicalTOMLOccurrencePath(
	k toml.Key,
	file FilePath,
	t reflect.Type,
	md *toml.MetaData,
	aotCurrent map[tomlPathKey]int,
	inlineCur map[tomlPathKey]int,
) (SchemaPath, OccurrencePath, OccurrenceKey, error) {
	st := tomlPathState{
		schema:    "",
		occ:       "",
		cur:       t,
		leaf:      OccurrenceKey(tomlLeaf(k)),
		inDynamic: false,
	}
	for i, seg := range k {
		prefixKey := tomlPathKey(strings.Join(k[:i+1], "."))

		done, err := tomlPathStep(&st, k, i, metaSeg(seg), prefixKey, file, md, aotCurrent, inlineCur)
		if err != nil {
			return "", "", "", err
		}

		if done {
			return st.schema, st.occ, st.leaf, nil
		}
	}

	return st.schema, st.occ, st.leaf, nil
}

type tomlPathState struct {
	schema    SchemaPath
	occ       OccurrencePath
	cur       reflect.Type
	leaf      OccurrenceKey
	inDynamic bool
}

func tomlLeaf(k toml.Key) metaSeg {
	if len(k) == 0 {
		return ""
	}

	return metaSeg(k[len(k)-1])
}

func tomlPathStep(
	st *tomlPathState,
	k toml.Key,
	i int,
	seg metaSeg,
	prefixKey tomlPathKey,
	file FilePath,
	md *toml.MetaData,
	aotCurrent map[tomlPathKey]int,
	inlineCur map[tomlPathKey]int,
) (bool, error) {
	st.cur = elemType(st.cur)

	switch {
	case st.inDynamic:
		tomlAppendSegment(st, seg, dynamicValueType(st.cur))
		st.inDynamic = isDynamicValueMap(st.cur)
	case st.cur.Kind() == reflect.Struct:
		done, err := tomlStepStruct(st, k, i, seg, file, md, aotCurrent, inlineCur)
		if err != nil || done {
			return done, err
		}
	case st.cur.Kind() == reflect.Map:
		tomlAppendSegment(st, seg, st.cur.Elem())
		st.inDynamic = false
	default:
		tomlAppendSegment(st, seg, st.cur)
	}

	applyTOMLArrayIndex(st, k, i, prefixKey, md, aotCurrent, inlineCur)

	return false, nil
}

func tomlAppendSegment(st *tomlPathState, seg metaSeg, nextType reflect.Type) {
	st.schema = schemaPath(st.schema, seg)
	st.occ = occPath(st.occ, seg)
	st.cur = nextType
}

func tomlStepStruct(
	st *tomlPathState,
	k toml.Key,
	i int,
	seg metaSeg,
	file FilePath,
	md *toml.MetaData,
	aotCurrent map[tomlPathKey]int,
	inlineCur map[tomlPathKey]int,
) (bool, error) {
	f, ok := fieldMatching(st.cur, keySegment(string(seg)))
	if ok {
		canonSeg := metaSeg(tagPathOf(f))
		st.schema = schemaPath(st.schema, canonSeg)
		st.occ = occPath(st.occ, canonSeg)
		st.cur = f.Type
		st.inDynamic = isDynamicMapField(f)

		return false, nil
	}

	if _, ok := extrasField(st.cur); ok {
		tomlAbsorbExtras(st, k, i, md, aotCurrent, inlineCur)

		return true, nil
	}

	return false, refuseAt(ErrUnknownField, file, 0,
		gorefuse.Subject(`"`+k.String()+`"`),
		gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
}
