package goconf

import (
	"reflect"
	"strings"
)

// presenceStep is one document segment translated into the schema's
// vocabulary: its canonical spelling, the type the next segment matches,
// and whether descent continues inside a dynamic map.
type presenceStep struct {
	next      reflect.Type
	canonical metaSeg
	dynamic   bool
}

// tagKey is one struct-tag key, e.g. "json".
type tagKey string

// declaredKey is a key spelled by a tag or a Go field name.
type declaredKey string

// formatTagKey is the struct-tag key one format decodes by.
func formatTagKey(format Format) tagKey {
	switch format {
	case FormatJSON:
		return tagKey("json")
	case FormatYAML:
		return tagKey("yaml")
	case FormatTOML:
		return tagKey("toml")
	case undetermined:
		return ""
	default:
		return ""
	}
}

// formatKeyEqual compares a document key with a declared name under one
// format's rules. JSON and TOML match names case-insensitively; YAML keys
// are exact, as the decoder requires.
func formatKeyEqual(format Format, declared declaredKey, seg keySegment) bool {
	if format == FormatYAML {
		return string(declared) == string(seg)
	}

	return strings.EqualFold(string(declared), string(seg))
}

// formatFieldMatching resolves a document segment using only its own format's
// tag (or Go name), so keys distinct in one format cannot collapse onto one
// field.
func formatFieldMatching(t reflect.Type, seg keySegment, format Format) (*reflect.StructField, bool) {
	t = elemType(t)
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	key := formatTagKey(format)

	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() || string(tagPathOf(f)) == "-" {
			continue
		}

		if fieldMatchesFormat(f, key, seg, format) {
			return f, true
		}
	}

	return nil, false
}

// fieldMatchesFormat compares one field's declared key for the source format,
// falling back to the Go name when that format declares no tag.
func fieldMatchesFormat(f *reflect.StructField, key tagKey, seg keySegment, format Format) bool {
	if key != "" {
		if tag, ok := f.Tag.Lookup(string(key)); ok {
			name, _, _ := strings.Cut(tag, ",")
			if name != "" {
				return formatKeyEqual(format, declaredKey(name), seg)
			}
		}
	}

	name := f.Name
	if format == FormatYAML {
		name = strings.ToLower(name)
	}

	return formatKeyEqual(format, declaredKey(name), seg)
}

// canonicalPresenceStep maps one segment through the decode target: matching
// uses the source format, canonical paths use tagPathOf. Extras keep document
// spelling; strict decode still refuses what no field declares.
func canonicalPresenceStep(t reflect.Type, inDynamic bool, seg pathSeg, format Format) presenceStep {
	if seg.isIdx {
		return presenceStep{canonical: "", next: elemType(t), dynamic: inDynamic}
	}

	t = elemType(t)
	if inDynamic {
		next := dynamicValueType(t)

		return presenceStep{canonical: seg.name, next: next, dynamic: isDynamicValueMap(next)}
	}

	if t.Kind() == reflect.Struct {
		f, ok := formatFieldMatching(t, keySegment(seg.name), format)
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
func canonicalOccurrencePaths(t reflect.Type, op OccurrencePath, format Format) (OccurrencePath, SchemaPath) {
	var outOp OccurrencePath

	var outSp SchemaPath

	cur := t
	dynamic := false
	rest := metaSeg(op)

	for rest != "" {
		seg, tail, hasTail := splitPathSeg(rest)
		step := canonicalPresenceStep(cur, dynamic, seg, format)

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

// canonicalizeMetaPaths translates one source's presence at the collector:
// matching uses the format, canonical paths stay schema-vocabulary. Occurrence.Key
// keeps document spelling and never drives defaults.
func canonicalizeMetaPaths(c *metaCollector, t reflect.Type, format Format) {
	keys := KeySet{paths: nil}

	for i := range c.entries {
		op, sp := canonicalOccurrencePaths(t, c.entries[i].Path, format)
		c.entries[i].Path = op
		c.entries[i].SchemaPath = sp
		keys = keys.Add(FieldPath(sp))
	}

	c.keys = keys
}
