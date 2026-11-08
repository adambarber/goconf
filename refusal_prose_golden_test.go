package goconf

// The golden fragments the prose tests pin.

// userSpan is one span a refusal must contain.
type userSpan string

// bannedVocab is internal vocabulary that must never reach a refusal.
type bannedVocab []string

// Shared golden fragments (goconst deduplicates the literals).
const (
	unknownFieldExtra   = userSpan(`unknown field: "extra"`)
	fixUnknownField     = userSpan("remove it from your config file, or the app must declare the field in its schema.")
	typeMismatchSpan    = userSpan("type mismatch")
	expectedNumberFound = userSpan("expected a number, found text")
	syntaxFixJSON       = userSpan("fix the JSON syntax in your config file")
	syntaxFixTOML       = userSpan("fix the TOML syntax in your config file")
	syntaxFixYAML       = userSpan("fix the YAML syntax in your config file")
)

// Banned internal vocabulary spans.
const (
	banDecodeTarget      = "decode target"
	banDecoderReported   = "the decoder reported"
	banDecoderRefusedTag = "decoder refused: JSON decode"
	banNotFoundInType    = "not found in type"
)

// proseCase is one refusal shape rendered through Load and judged.
type proseCase struct {
	body     SourceBytes
	name     caseName
	format   Format
	contains []userSpan
	excludes bannedVocab
}

// unknownFieldCase is the shared unknown-field row per format.
func unknownFieldCase(name caseName, format Format, body SourceBytes) proseCase {
	return proseCase{
		body:     body,
		name:     name,
		format:   format,
		contains: []userSpan{unknownFieldExtra, fixUnknownField},
		excludes: bannedVocab{banDecodeTarget, banNotFoundInType},
	}
}
