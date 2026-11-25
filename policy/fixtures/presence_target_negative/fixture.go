// The near miss: the collector forwards the target type it was given.
package fixtures

import "reflect"

type metaCollector struct{}

func canonicalizeMetaPaths(c *metaCollector, t reflect.Type) {}

func collectFixtureMeta(src []byte, sourceIndex int, t reflect.Type) (bool, []string, bool) {
	c := &metaCollector{}
	canonicalizeMetaPaths(c, t)
	_ = src
	_ = sourceIndex

	return len(src) > 0, []string{string(src)}, len(src) == 0
}
