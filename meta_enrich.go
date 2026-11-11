package goconf

import (
	"reflect"
	"strconv"
	"strings"
)

type pathSeg struct {
	name  metaSeg
	index int
	isIdx bool
}

// resolvePathValue resolves an indexed occurrence path against a reflected
// config value.
func resolvePathValue(root reflect.Value, path OccurrencePath) (reflect.Value, bool) {
	v := derefReflect(root)
	if path == "" {
		return v, v.IsValid()
	}

	rest := metaSeg(path)
	for rest != "" {
		seg, tail, hasTail := splitPathSeg(rest)

		next, ok := stepPathValue(v, seg)
		if !ok {
			return reflect.Value{}, false
		}

		v = next

		if !hasTail {
			return v, true
		}

		rest = tail
	}

	return v, v.IsValid()
}

// lookupEffectiveValue resolves the value at an indexed occurrence path in the
// final decoded config. The bool reports whether the path resolved.
func lookupEffectiveValue(root reflect.Value, path OccurrencePath) (reflect.Value, bool) {
	if path == "" {
		return root, root.IsValid()
	}

	v, ok := resolvePathValue(root, path)
	if !ok || !v.IsValid() {
		return reflect.Value{}, false
	}

	return v, true
}

func splitPathSeg(path metaSeg) (pathSeg, metaSeg, bool) {
	if strings.HasPrefix(string(path), "[") {
		return splitBracketSeg(path)
	}

	return splitNamedSeg(path)
}

func splitBracketSeg(path metaSeg) (pathSeg, metaSeg, bool) {
	s := string(path)

	end := strings.IndexByte(s, ']')
	if end < 0 {
		return pathSeg{name: path, index: 0, isIdx: false}, "", false
	}

	n, err := strconv.Atoi(s[1:end])
	tail := trimDot(metaSeg(s[end+1:]))
	hasTail := tail != ""

	if err != nil {
		return pathSeg{name: metaSeg(s[:end+1]), index: 0, isIdx: false}, tail, hasTail
	}

	return pathSeg{name: "", index: n, isIdx: true}, tail, hasTail
}

func splitNamedSeg(path metaSeg) (pathSeg, metaSeg, bool) {
	s := string(path)
	dot := strings.IndexByte(s, '.')
	br := strings.IndexByte(s, '[')
	cut := segCut(dot, br)

	if cut < 0 {
		return pathSeg{name: path, index: 0, isIdx: false}, "", false
	}

	name := metaSeg(s[:cut])
	if s[cut] == '.' {
		tail := metaSeg(s[cut+1:])

		return pathSeg{name: name, index: 0, isIdx: false}, tail, tail != ""
	}

	tail := metaSeg(s[cut:])

	return pathSeg{name: name, index: 0, isIdx: false}, tail, tail != ""
}

func segCut(dot, br int) int {
	switch {
	case dot >= 0 && br >= 0:
		if dot < br {
			return dot
		}

		return br
	case dot >= 0:
		return dot
	case br >= 0:
		return br
	default:
		return -1
	}
}

func trimDot(s metaSeg) metaSeg { return metaSeg(strings.TrimPrefix(string(s), ".")) }
