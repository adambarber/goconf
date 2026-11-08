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
	expectedNumberFound = userSpan("is text, not a whole number")
	typeMismatchKeySpan = userSpan(`the value of "retries"`)
	mismatchFixRetries  = userSpan(`change the value of "retries" to a whole number`)
	syntaxFixJSON       = userSpan("check the last line for a missing bracket or quote")
	syntaxFixTOML       = userSpan("check the last line for a missing bracket or quote")
	syntaxFixYAML       = userSpan("check the line for a missing bracket, quote, or indentation")
)

// Banned internal vocabulary spans.
const (
	banDecodeTarget      = "decode target"
	banDecoderReported   = "the decoder reported"
	banDecoderRefusedTag = "decoder refused: JSON decode"
	banNotFoundInType    = "not found in type"
	banDecoderRefused    = "decoder refused"
	// banFixTheErrorTheDecoderReported is the circular fix A4 bans.
	banFixTheErrorTheDecoderReported = "fix the error the decoder reported"
)

// proseCase is one refusal shape rendered through Load and judged.
type proseCase struct {
	contains []userSpan
	excludes bannedVocab
	body     SourceBytes
	name     caseName
	format   Format
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
