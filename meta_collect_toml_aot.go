package goconf

import (
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

func emitTOMLAOTContainers(c *metaCollector, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, k toml.Key, line int) error {
	if len(k) == 0 || md.Type(k...) != tomlTypeArrayHash {
		return nil
	}

	arrayKey := tomlPathKey(strings.Join(k, "."))
	if st.aotContAdded[arrayKey] {
		return nil
	}

	st.aotContAdded[arrayKey] = true

	schema, occ, _, err := canonicalTOMLOccurrencePath(k, file, t, md, st.aotCurrent, st.inlineCur)
	if err != nil {
		return err
	}

	contOcc, contSchema, contKey := tomlAOTContainerFromElement(occ, schema, k)
	val := tomlAOTContainerValue(raw, md, k, st.aotCurrent)
	c.add(contOcc, contSchema, contKey, line, val)

	return nil
}

func tomlAOTContainerFromElement(occ OccurrencePath, schema SchemaPath, k toml.Key) (OccurrencePath, SchemaPath, OccurrenceKey) {
	name := k[len(k)-1]
	key := OccurrenceKey(name)

	// element path ends with [i]; container is path without trailing index for this AOT segment
	contOcc := stripTrailingAOTIndex(occ)

	// schema path for AOT element may be same as container schema in our model
	return contOcc, schema, key
}

func stripTrailingAOTIndex(occ OccurrencePath) OccurrencePath {
	s := string(occ)

	trimmed, ok := strings.CutSuffix(s, "]")
	if !ok {
		return occ
	}

	i := len(trimmed) - 1
	for i >= 0 && trimmed[i] != '[' {
		i--
	}

	if i < 0 {
		return occ
	}

	return OccurrencePath(trimmed[:i])
}

func tomlAOTContainerValue(raw tomlKeyMap, md *toml.MetaData, k toml.Key, aotCurrent map[tomlPathKey]int) reflect.Value {
	// best-effort: container value is the array/slice at this path in decoded structure
	// not always representable from primitive map without full decode; enrichment may help but
	// containers are structural. use lookup of the array path where possible.
	_ = raw
	_ = md
	_ = k
	_ = aotCurrent

	return reflect.Value{}
}

func emitTOMLInlineArrayElements(c *metaCollector, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, k toml.Key, line int) error {
	if !tomlIsScalarInlineArrayKey(md, k) {
		return nil
	}

	arr := derefReflect(tomlLookupMetaValue(raw, md, k, st.aotCurrent, st.inlineCur))
	if !arr.IsValid() {
		return nil
	}

	if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
		return nil
	}

	schema, occ, _, pathErr := canonicalTOMLOccurrencePath(k, file, t, md, st.aotCurrent, st.inlineCur)
	if pathErr != nil {
		return pathErr
	}

	tomlAddIndexedElemRange(c, occ, schema, arr, line)

	return nil
}

func tomlAddIndexedElemRange(c *metaCollector, occ OccurrencePath, schema SchemaPath, arr reflect.Value, line int) {
	for i := range arr.Len() {
		tomlAddIndexedElem(c, occ, schema, i, line, derefReflect(arr.Index(i)))
	}
}

func tomlIsScalarInlineArrayKey(md *toml.MetaData, k toml.Key) bool {
	if len(k) == 0 || md.Type(k...) != tomlTypeArray {
		return false
	}

	return !tomlInlineArrayHasExpandedKeys(md, k)
}
