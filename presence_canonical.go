package goconf

import "reflect"

// presenceStep is one document segment translated into the schema's
// vocabulary: its canonical spelling, the type the next segment matches,
// and whether descent continues inside a dynamic map.
type presenceStep struct {
	next      reflect.Type
	canonical metaSeg
	dynamic   bool
}

// canonicalPresenceStep maps one document segment through the decode target,
// reusing the field vocabulary TOML walks. Extras and unknown keys keep their
// document spelling; strict decode still refuses what no field declares.
func canonicalPresenceStep(t reflect.Type, inDynamic bool, seg pathSeg) presenceStep {
	if seg.isIdx {
		return presenceStep{canonical: "", next: elemType(t), dynamic: inDynamic}
	}

	t = elemType(t)
	if inDynamic {
		next := dynamicValueType(t)

		return presenceStep{canonical: seg.name, next: next, dynamic: isDynamicValueMap(next)}
	}

	if t.Kind() == reflect.Struct {
		f, ok := fieldMatching(t, keySegment(seg.name))
		if ok {
			return presenceStep{canonical: metaSeg(tagPathOf(f)), next: f.Type, dynamic: isDynamicMapField(f)}
		}

		return presenceStep{canonical: seg.name, next: t, dynamic: false}
	}

	if t.Kind() == reflect.Map {
		return presenceStep{canonical: seg.name, next: t.Elem(), dynamic: false}
	}

	return presenceStep{canonical: seg.name, next: t, dynamic: false}
}

// canonicalOccurrencePaths rewrites an indexed path into schema spelling and
// derives its unindexed schema path from the same translated segments.
func canonicalOccurrencePaths(t reflect.Type, op OccurrencePath) (OccurrencePath, SchemaPath) {
	var outOp OccurrencePath

	var outSp SchemaPath

	cur := t
	dynamic := false
	rest := metaSeg(op)

	for rest != "" {
		seg, tail, hasTail := splitPathSeg(rest)
		step := canonicalPresenceStep(cur, dynamic, seg)

		if seg.isIdx {
			outOp = occIndex(outOp, seg.index)
		} else {
			outOp = occPath(outOp, step.canonical)
			outSp = schemaPath(outSp, step.canonical)
		}

		cur = step.next
		dynamic = step.dynamic

		if !hasTail {
			break
		}

		rest = tail
	}

	return outOp, outSp
}

// canonicalizeMetaPaths translates one source's presence set and occurrences
// once, at the collector. Occurrence.Key keeps the document spelling as
// metadata; it never drives defaults.
func canonicalizeMetaPaths(c *metaCollector, t reflect.Type) {
	keys := KeySet{paths: nil}

	for i := range c.entries {
		op, sp := canonicalOccurrencePaths(t, c.entries[i].Path)
		c.entries[i].Path = op
		c.entries[i].SchemaPath = sp
		keys = keys.Add(FieldPath(sp))
	}

	c.keys = keys
}
