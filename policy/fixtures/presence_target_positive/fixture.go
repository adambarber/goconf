// The defect: a presence collector holds the decode target type but passes a
// fabricated one to the single translation helper.
package fixtures

import "reflect"

type Format string

const FormatFixture Format = "fixture"

type metaCollector struct{}

func canonicalizeMetaPaths(c *metaCollector, t reflect.Type, f Format) {}

func collectFixtureMeta(src []byte, sourceIndex int, t reflect.Type) (bool, []string, bool) {
	c := &metaCollector{}
	canonicalizeMetaPaths(c, reflect.TypeFor[int](), FormatFixture)
	_ = src
	_ = sourceIndex
	_ = t

	return len(src) > 0, []string{string(src)}, len(src) == 0
}
