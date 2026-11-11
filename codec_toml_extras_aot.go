package goconf

import (
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

type tomlAOTIndex map[tomlPathKey]int

func tomlUndecodedAOTIndices(undec []toml.Key, md *toml.MetaData) []tomlAOTIndex {
	out := make([]tomlAOTIndex, len(undec))
	if len(undec) == 0 {
		return out
	}

	aotNext := map[tomlPathKey]int{}
	aotCurrent := map[tomlPathKey]int{}
	ui := 0

	for _, k := range md.Keys() {
		updateAOTContext(k, md, aotNext, aotCurrent)

		if ui < len(undec) && tomlKeysEqual(k, undec[ui]) {
			out[ui] = copyTOMLAOTIndex(aotCurrent)
			ui++

			if ui == len(undec) {
				break
			}
		}
	}

	return out
}

func tomlKeysEqual(a, b toml.Key) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func copyTOMLAOTIndex(src map[tomlPathKey]int) tomlAOTIndex {
	if len(src) == 0 {
		return tomlAOTIndex{}
	}

	dst := make(tomlAOTIndex, len(src))
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

// tomlCaptureAOTIndexed descends into the AOT element selected for the
// array-of-tables path prefix k[:nextIdx] when that prefix has a current
// element index for this undecoded occurrence.
func tomlCaptureAOTIndexed(v, raw reflect.Value, k toml.Key, aot tomlAOTIndex, nextIdx int, file FilePath) (bool, bool, error) {
	elemIdx, ok := tomlAOTCaptureIndex(v, raw, k, aot, nextIdx)
	if !ok {
		return false, false, nil
	}

	captured, err := tomlCaptureAt(v.Index(elemIdx), raw.Index(elemIdx), k, aot, nextIdx, file)

	return true, captured, err
}

func tomlAOTCaptureIndex(v, raw reflect.Value, k toml.Key, aot tomlAOTIndex, nextIdx int) (int, bool) {
	v, raw, ok := tomlAOTCaptureValues(v, raw, nextIdx, k)
	if !ok {
		return 0, false
	}

	prefix := tomlPathKey(strings.Join(k[:nextIdx], "."))

	elemIdx, ok := aot[prefix]
	if !ok {
		return 0, false
	}

	if elemIdx < 0 || elemIdx >= v.Len() || elemIdx >= raw.Len() {
		return 0, false
	}

	return elemIdx, true
}

func tomlAOTCaptureValues(v, raw reflect.Value, nextIdx int, k toml.Key) (reflect.Value, reflect.Value, bool) {
	if nextIdx > len(k) {
		return reflect.Value{}, reflect.Value{}, false
	}

	v = derefValue(v)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return reflect.Value{}, reflect.Value{}, false
	}

	if raw.Kind() == reflect.Interface && !raw.IsNil() {
		raw = raw.Elem()
	}

	if raw.Kind() != reflect.Slice && raw.Kind() != reflect.Array {
		return reflect.Value{}, reflect.Value{}, false
	}

	return v, raw, true
}
