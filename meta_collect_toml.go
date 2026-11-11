package goconf

import (
	"bytes"
	"reflect"

	"github.com/BurntSushi/toml"
)

const (
	tomlTypeArrayHash = "ArrayHash"
	tomlTypeArray     = "Array"
)

type tomlPathKey string

func collectTOMLMeta(src SourceBytes, file FilePath, sourceIndex int, t reflect.Type) (KeySet, []Occurrence, error) {
	c := newMetaCollector(file, sourceIndex)

	var raw tomlKeyMap

	md, err := toml.NewDecoder(bytes.NewReader(src)).Decode(&raw)
	if err != nil {
		return KeySet{}, nil, tomlRefusal(file, errTOMLValueFree(err))
	}

	if err := emitTOMLKeys(c, src, file, t, raw, &md); err != nil {
		return KeySet{}, nil, err
	}

	keys, entries := c.result()

	return keys, entries, nil
}

type tomlCollectState struct {
	aotNext         map[tomlPathKey]int
	aotCurrent      map[tomlPathKey]int
	inlineCur       map[tomlPathKey]int
	inlineSeen      map[tomlPathKey]map[tomlPathKey]bool
	inlineElemAdded map[tomlPathKey]map[int]bool
	aotContAdded    map[tomlPathKey]bool
	cursor          int
}

func newTOMLCollectState() *tomlCollectState {
	return &tomlCollectState{
		cursor:          0,
		aotNext:         map[tomlPathKey]int{},
		aotCurrent:      map[tomlPathKey]int{},
		inlineCur:       map[tomlPathKey]int{},
		inlineSeen:      map[tomlPathKey]map[tomlPathKey]bool{},
		inlineElemAdded: map[tomlPathKey]map[int]bool{},
		aotContAdded:    map[tomlPathKey]bool{},
	}
}

func emitTOMLKeys(c *metaCollector, src SourceBytes, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData) error {
	st := newTOMLCollectState()

	for _, k := range md.Keys() {
		if err := emitTOMLKeyEntry(c, src, file, t, raw, md, st, k); err != nil {
			return err
		}
	}

	return nil
}

func emitTOMLKeyEntry(c *metaCollector, src SourceBytes, file FilePath, t reflect.Type, raw tomlKeyMap, md *toml.MetaData, st *tomlCollectState, k toml.Key) error {
	line, next := tomlFindKeyLine(src, st.cursor, k, md)
	if next > st.cursor {
		st.cursor = next
	}

	updateAOTContext(k, md, st.aotNext, st.aotCurrent)
	updateInlineContext(k, md, st)

	steps := []func() error{
		func() error { return emitTOMLAOTContainers(c, file, t, raw, md, st, k, line) },
		func() error { return emitTOMLInlineTableArrayContainers(c, file, t, raw, md, st, k, line) },
		func() error { return emitTOMLKeyOccurrence(c, file, t, raw, md, st, k, line) },
		func() error { return emitTOMLInlineArrayElements(c, file, t, raw, md, st, k, line) },
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}
