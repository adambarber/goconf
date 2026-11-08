package goconf

import "github.com/adambarber/gorefuse"

// Reason is why goconf refused a load: one domain-typed identity a caller
// extracts with errors.As into gorefuse.Because[Reason].
type Reason string

// Error makes every Reason an error, so a caller can errors.Is one
// Because's reason against another's.
func (r Reason) Error() string { return string(r) }

// refusalError carries one already-stripped refusal text as the decoder
// refusal's cause: the identity comes from the Reason, never from here.
// The text must ride as text — passing the underlying decoder error would
// echo config values and decoder vocabulary the strip passes exist to
// remove — and the Cause slot takes an error, not a string. A named type
// (not errors.New) because err113 forbids dynamic errors and no nolint is
// permitted; this one string-holder is the documented exception.
func refusalError(msg refusalText) error { return refusalCauseError{msg: msg} }

// refusalCauseError is refusalError's concrete error type.
type refusalCauseError struct{ msg refusalText }

// The refusal reasons goconf returns, one per distinct refusal semantics;
// near-identical per-format sentinels share one Reason.
const (
	ErrUnknownFormat    Reason = "unknown format"
	ErrAmbiguous        Reason = "ambiguous candidates"
	ErrUnknownCandidate Reason = "unknown candidate"
	ErrEscapingName     Reason = "escaping config name"
	ErrOutsideSearch    Reason = "outside search paths"
	ErrDuplicateKey     Reason = "duplicate key"
	ErrMultiDoc         Reason = "multi-document stream"
	ErrNotAMapping      Reason = "not a mapping"
	ErrDecoderRefused   Reason = "decoder refused"
	ErrIO               Reason = "io fault"
	ErrUnknownField     Reason = "unknown field"
	ErrTypeMismatch     Reason = "type mismatch"
	ErrNothingFound     Reason = "no config file found"
	ErrUnmarshalable    Reason = "unmarshalable default"
	ErrBadDefaultShape  Reason = "bad default shape"
	ErrTargetCast       Reason = "target cast failure"
)

// Error renders the stripped refusal message: the Cause tail gorefuse
// carries natively — see Because.Cause and the causeTail render stage.
func (e refusalCauseError) Error() string { return string(e.msg) }

// refuseAt builds the refusal for a rejection that names one file and
// line: the frame points at the file that was refused, never at goconf.
func refuseAt(r Reason, file FilePath, line gorefuse.Line, subject gorefuse.Subject, why gorefuse.Why) error {
	return gorefuse.Because[Reason]{
		Reason: r,
		Cause:  nil,
		What:   gorefuse.At(gorefuse.File(file), line, subject, why),
	}
}

// refuseIO builds an IO-fault refusal carrying the underlying error as
// its cause tail: the detail survives, the identity is ErrIO.
func refuseIO(what gorefuse.Refusal, cause error) error {
	return gorefuse.Because[Reason]{
		Reason: ErrIO,
		Cause:  cause,
		What:   what,
	}
}

// refuseDecoder builds the decoder-refused refusal carrying the
// underlying decoder's error as its cause tail: the detail survives, the
// identity is ErrDecoderRefused.
func refuseDecoder(what gorefuse.Refusal, cause error) error {
	return refuseReason(ErrDecoderRefused, what, cause)
}

// refuseReason builds a refusal with an explicit Reason carrying the
// underlying error as its cause tail.
func refuseReason(r Reason, what gorefuse.Refusal, cause error) error {
	return gorefuse.Because[Reason]{
		Reason: r,
		Cause:  cause,
		What:   what,
	}
}

// refuseWhy builds the frameless refusal for a discovery- or option-stage
// rejection: no file was refused, so no frame is fabricated.
func refuseWhy(r Reason, subject gorefuse.Subject, why gorefuse.Why) error {
	return gorefuse.Because[Reason]{
		Reason: r,
		Cause:  nil,
		What:   gorefuse.At("", 0, subject, why),
	}
}
