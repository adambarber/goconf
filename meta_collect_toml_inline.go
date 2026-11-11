package goconf

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

func updateAOTContext(k toml.Key, md *toml.MetaData, aotNext, aotCurrent map[tomlPathKey]int) {
	for i := 1; i <= len(k); i++ {
		prefix := k[:i]
		if md.Type(prefix...) != tomlTypeArrayHash {
			continue
		}

		if i != len(k) {
			continue
		}

		keyPath := tomlPathKey(strings.Join(prefix, "."))
		aotCurrent[keyPath] = aotNext[keyPath]
		aotNext[keyPath]++
		clearNestedAOT(keyPath, aotCurrent, aotNext)
	}
}

func clearNestedAOT(prefix tomlPathKey, aotCurrent, aotNext map[tomlPathKey]int) {
	p := string(prefix) + "."
	for k := range aotCurrent {
		if !strings.HasPrefix(string(k), p) {
			continue
		}

		delete(aotCurrent, k)
		delete(aotNext, k)
	}
}

func updateInlineContext(k toml.Key, md *toml.MetaData, st *tomlCollectState) {
	for i := 1; i <= len(k); i++ {
		prefix := k[:i]
		if md.Type(prefix...) != tomlTypeArray || i == len(k) {
			continue
		}

		arrayKey := tomlPathKey(strings.Join(prefix, "."))
		suffix := tomlPathKey(strings.Join(k[i:], "."))
		tomlAdvanceInlineElement(st, arrayKey, suffix)
	}
}

func tomlAdvanceInlineElement(st *tomlCollectState, arrayKey, suffix tomlPathKey) {
	seen, ok := st.inlineSeen[arrayKey]
	if !ok {
		st.inlineSeen[arrayKey] = map[tomlPathKey]bool{suffix: true}
		st.inlineCur[arrayKey] = 0

		return
	}

	if !seen[suffix] {
		seen[suffix] = true

		return
	}

	st.inlineCur[arrayKey]++
	st.inlineSeen[arrayKey] = map[tomlPathKey]bool{suffix: true}
}

func emitTOMLKeyOccurrence(c *metaCollector, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, k toml.Key, line int) error {
	schema, occ, leaf, err := canonicalTOMLOccurrencePath(k, file, t, md, st.aotCurrent, st.inlineCur)
	if err != nil {
		return err
	}

	val := tomlLookupMetaValue(raw, md, k, st.aotCurrent, st.inlineCur)

	if md.Type(k...) == tomlTypeArrayHash {
		idx, ok := st.aotCurrent[tomlPathKey(strings.Join(k, "."))]
		if ok {
			leaf = OccurrenceKey("[" + strconv.Itoa(idx) + "]")
		}
	}

	c.add(occ, schema, leaf, line, val)

	return nil
}

func emitTOMLInlineTableArrayContainers(c *metaCollector, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, k toml.Key, line int) error {
	for i := 1; i < len(k); i++ {
		prefix := k[:i]
		if !tomlIsExpandedInlineArrayPrefix(md, prefix) {
			continue
		}

		if err := emitTOMLInlineContainerAt(c, file, t, raw, md, st, prefix, line); err != nil {
			return err
		}
	}

	return nil
}

func tomlIsExpandedInlineArrayPrefix(md *toml.MetaData, prefix toml.Key) bool {
	if md.Type(prefix...) != tomlTypeArray {
		return false
	}

	return tomlInlineArrayHasExpandedKeys(md, prefix)
}

func emitTOMLInlineContainerAt(c *metaCollector, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, prefix toml.Key, line int) error {
	arrayKey := tomlPathKey(strings.Join(prefix, "."))
	idx := st.inlineCur[arrayKey]

	if tomlInlineElemAlreadyAdded(st, arrayKey, idx) {
		return nil
	}

	schema, occ, _, pathErr := canonicalTOMLOccurrencePath(prefix, file, t, md, st.aotCurrent, st.inlineCur)
	if pathErr == nil {
		tomlAddIndexedElem(c, occ, schema, idx, line, tomlInlineElemValue(raw, md, st, prefix, idx))
	}

	return pathErr
}

func tomlAddIndexedElem(c *metaCollector, occ OccurrencePath, schema SchemaPath, idx, line int, val reflect.Value) {
	c.add(occIndex(occ, idx), schemaIndex(schema), OccurrenceKey("["+strconv.Itoa(idx)+"]"), line, val)
}

func tomlInlineElemAlreadyAdded(st *tomlCollectState, arrayKey tomlPathKey, idx int) bool {
	added := st.inlineElemAdded[arrayKey]
	if added == nil {
		added = map[int]bool{}
		st.inlineElemAdded[arrayKey] = added
	}

	if added[idx] {
		return true
	}

	added[idx] = true

	return false
}

func tomlInlineElemValue(raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, prefix toml.Key, idx int) reflect.Value {
	arr := derefReflect(tomlLookupMetaValue(raw, md, prefix, st.aotCurrent, st.inlineCur))
	if !arr.IsValid() || (arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array) {
		return reflect.Value{}
	}

	if idx < 0 || idx >= arr.Len() {
		return reflect.Value{}
	}

	return derefReflect(arr.Index(idx))
}

func tomlInlineArrayHasExpandedKeys(md *toml.MetaData, k toml.Key) bool {
	prefix := strings.Join(k, ".") + "."
	for _, other := range md.Keys() {
		if strings.HasPrefix(strings.Join(other, "."), prefix) {
			return true
		}
	}

	return false
}
