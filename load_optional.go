package goconf

import "log"

// warnOptionalAbsence prints the absence warning once per Load call, at
// the point true absence is accepted: the caller opted in, but a silent
// empty config would hide the fact that nothing was loaded.
func warnOptionalAbsence() {
	log.Print("[warn] No config file found. Using default values.")
}
