package goconf

// Layers is one merged configuration view: several sources decoded in
// order, later sources overriding earlier ones field by field. It answers
// "was this field set by any source" so defaults can fill only absences.
type Layers struct {
	present KeySet
	sources []SourceBytes
}

// Present reports the field paths set by any source, including explicit
// nulls: a null is a decision and counts as present.
func (l Layers) Present() KeySet {
	return l.present
}

// layerSources composes sources in order, later overriding earlier. Every
// source must decode cleanly against the same format.
func layerSources(sources []SourceBytes) (Layers, error) {
	l := Layers{present: KeySet{paths: nil}, sources: sources}
	for _, src := range sources {
		keys, err := presentJSONKeys(src)
		if err != nil {
			return l, err
		}

		for _, p := range keys.pathsList() {
			l.present = l.present.Add(p)
		}
	}

	return l, nil
}

// decodeMerged decodes the sources in order into one target. A later
// source only rewrites the fields it sets, so nesting deep-merges.
func decodeMerged(l Layers, dst Target) error {
	for _, src := range l.sources {
		if err := decodeJSONStrict(src, dst); err != nil {
			return err
		}
	}

	return nil
}
