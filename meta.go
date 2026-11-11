package goconf

import "reflect"

// OccurrencePath identifies one specific key occurrence by indexed path.
type OccurrencePath string

// SchemaPath identifies the unindexed schema path used by presence/refusals.
type SchemaPath FieldPath

// OccurrenceKey is the leaf key or element name at an occurrence.
type OccurrenceKey string

// Occurrence is one per-occurrence record of a present key/value in a source.
// Records are in document order per source; sources follow load order.
type Occurrence struct {
	// Value holds the occurrence value (effective when resolvable, else source).
	Value reflect.Value
	// File is the source file for this occurrence.
	File FilePath
	// Key is the leaf key or element name for this occurrence.
	Key OccurrenceKey
	// Path is the indexed occurrence path for this occurrence.
	Path OccurrencePath
	// SchemaPath is the unindexed canonical schema path.
	SchemaPath SchemaPath
	// SourceIndex is the contributing source/provenance index.
	SourceIndex int
	// Line is the 1-based source line of this occurrence.
	Line int
	// Effective reports contribution to the effective loaded config.
	Effective bool
}

// Meta is an ordered, queryable collection of occurrence records.
type Meta struct {
	entries []Occurrence
}

// Entries returns all recorded occurrences in load/document order.
func (m *Meta) Entries() []Occurrence {
	if m == nil {
		return nil
	}

	out := make([]Occurrence, len(m.entries))
	copy(out, m.entries)

	return out
}

// Len reports the number of recorded occurrences.
func (m *Meta) Len() int {
	if m == nil {
		return 0
	}

	return len(m.entries)
}

// ByPath returns occurrences matching an indexed occurrence path.
func (m *Meta) ByPath(path OccurrencePath) []Occurrence {
	return m.filter(func(o Occurrence) bool { return o.Path == path })
}

// BySchemaPath returns occurrences matching an unindexed schema path.
func (m *Meta) BySchemaPath(path SchemaPath) []Occurrence {
	return m.filter(func(o Occurrence) bool { return o.SchemaPath == path })
}

// Lookup is equivalent to ByPath.
func (m *Meta) Lookup(path OccurrencePath) []Occurrence { return m.ByPath(path) }

// LookupSchema is equivalent to BySchemaPath.
func (m *Meta) LookupSchema(path SchemaPath) []Occurrence { return m.BySchemaPath(path) }

// At returns the indexed-path occurrence at position index.
func (m *Meta) At(path OccurrencePath, index int) (Occurrence, bool) {
	return metaAt(m.ByPath(path), index)
}

// AtSchema returns the schema-path occurrence at position index.
func (m *Meta) AtSchema(path SchemaPath, index int) (Occurrence, bool) {
	return metaAt(m.BySchemaPath(path), index)
}

func metaAt(matches []Occurrence, index int) (Occurrence, bool) {
	if index < 0 || index >= len(matches) {
		return Occurrence{
			Value:       reflect.Value{},
			File:        "",
			Key:         "",
			Path:        "",
			SchemaPath:  "",
			SourceIndex: 0,
			Line:        0,
			Effective:   false,
		}, false
	}

	return matches[index], true
}

// Effective returns occurrences contributing to the effective config.
func (m *Meta) Effective() []Occurrence {
	return m.filter(func(o Occurrence) bool { return o.Effective })
}

func (m *Meta) filter(keep func(Occurrence) bool) []Occurrence {
	if m == nil {
		return nil
	}

	out := []Occurrence{}

	for _, e := range m.entries {
		if keep(e) {
			out = append(out, e)
		}
	}

	return out
}

// metaCollector builds per-occurrence metadata and presence for one source.
type metaCollector struct {
	keys        KeySet
	file        FilePath
	entries     []Occurrence
	sourceIndex int
}

func newMetaCollector(file FilePath, sourceIndex int) *metaCollector {
	return &metaCollector{
		keys:        KeySet{paths: nil},
		entries:     nil,
		file:        file,
		sourceIndex: sourceIndex,
	}
}

func (c *metaCollector) add(path OccurrencePath, schema SchemaPath, key OccurrenceKey, line int, value reflect.Value) {
	c.keys = c.keys.Add(FieldPath(schema))
	c.entries = append(c.entries, Occurrence{
		Value:       value,
		File:        c.file,
		Key:         key,
		Path:        path,
		SchemaPath:  schema,
		SourceIndex: c.sourceIndex,
		Line:        line,
		Effective:   false,
	})
}

func (c *metaCollector) result() (KeySet, []Occurrence) { return c.keys, c.entries }
