// The near miss: the collector forwards the target type it was given.
package fixtures

import "reflect"

type Format string

const FormatFixture Format = "fixture"

type metaCollector struct{}

func canonicalizeMetaPaths(c *metaCollector, t reflect.Type, f Format) {}

func collectFixtureMeta(src []byte, sourceIndex int, t reflect.Type) (bool, []string, bool) {
	c := &metaCollector{}
	canonicalizeMetaPaths(c, t, FormatFixture)
	_ = src
	_ = sourceIndex

	return len(src) > 0, []string{string(src)}, len(src) == 0
}
