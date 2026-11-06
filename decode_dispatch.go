package goconf

// decodeFormat decodes one source's bytes into the target through the
// format's strict decoder. Format is a closed enum, so the switch is
// total over its constants.
func decodeFormat(src SourceBytes, dst Target, f Format) error {
	switch f {
	case FormatJSON:
		return decodeJSONStrict(src, dst)
	case FormatYAML:
		return decodeYAMLStrict(src, dst)
	case FormatTOML:
		return decodeTOMLStrict(src, dst)
	case undetermined:
		return errUnknownFormat
	default:
		return errUnknownFormat
	}
}

// presenceOf records the field paths present in one source, per format.
func presenceOf(src SourceBytes, f Format) (KeySet, error) {
	switch f {
	case FormatJSON:
		return presentJSONKeys(src)
	case FormatYAML:
		return presentYAMLKeys(src)
	case FormatTOML:
		return presentTOMLKeys(src)
	case undetermined:
		return KeySet{}, errUnknownFormat
	default:
		return KeySet{}, errUnknownFormat
	}
}

// unionPresence merges the presence views of all sources, later sources
// adding to earlier ones.
func unionPresence(sources []SourceBytes, f Format) (KeySet, error) {
	union := KeySet{paths: nil}

	for _, src := range sources {
		keys, err := presenceOf(src, f)
		if err != nil {
			return union, err
		}

		for _, p := range keys.pathsList() {
			union = union.Add(p)
		}
	}

	return union, nil
}
