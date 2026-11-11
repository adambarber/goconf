package goconf

import (
	"strings"

	"github.com/BurntSushi/toml"
)

func tomlAbsorbExtras(st *tomlPathState, k toml.Key, start int, md *toml.MetaData, aotCurrent, inlineCur map[tomlPathKey]int) {
	for j := start; j < len(k); j++ {
		st.schema = schemaPath(st.schema, metaSeg(k[j]))
		st.occ = occPath(st.occ, metaSeg(k[j]))

		pj := tomlPathKey(strings.Join(k[:j+1], "."))
		prefix := k[:j+1]
		applyExtrasArrayIndex(st, prefix, pj, j < len(k)-1, md, aotCurrent, inlineCur)
	}
}

func applyExtrasArrayIndex(st *tomlPathState, prefix toml.Key, pj tomlPathKey, hasChild bool, md *toml.MetaData, aotCurrent, inlineCur map[tomlPathKey]int) {
	if md.Type(prefix...) == tomlTypeArrayHash {
		if idx, ok := aotCurrent[pj]; ok {
			st.occ = occIndex(st.occ, idx)
		}

		return
	}

	if hasChild && md.Type(prefix...) == tomlTypeArray {
		if idx, ok := inlineCur[pj]; ok {
			st.occ = occIndex(st.occ, idx)
		}
	}
}

func applyTOMLArrayIndex(st *tomlPathState, k toml.Key, i int, prefixKey tomlPathKey, md *toml.MetaData, aotCurrent, inlineCur map[tomlPathKey]int) {
	prefix := k[:i+1]
	typ := md.Type(prefix...)

	if typ == tomlTypeArrayHash {
		if idx, ok := aotCurrent[prefixKey]; ok {
			st.occ = occIndex(st.occ, idx)
		}

		st.cur = elemType(st.cur)

		return
	}

	if typ == tomlTypeArray && i < len(k)-1 {
		if idx, ok := inlineCur[prefixKey]; ok {
			st.occ = occIndex(st.occ, idx)
		}

		st.cur = elemType(st.cur)
	}
}
