package goconf

import "github.com/adambarber/gorefuse"

// byteCap is one source's size limit, in bytes.
type byteCap int64

const (
	defaultByteCap byteCap = 16 << 20
	hardByteCap    byteCap = 64 << 20
)

// WithMaxBytes caps each source at n bytes for one load, overriding the
// default of 16 MiB. n must be positive and no greater than the hard maximum
// of 64 MiB; an out-of-range n is a programmer error and panics.
func WithMaxBytes(n int) *valueOption[byteCap] {
	if n < 1 || byteCap(n) > hardByteCap {
		panic("goconf: WithMaxBytes must be between 1 and 64 MiB")
	}

	return newValueOption(byteCap(n), func(s *settings, capBytes byteCap) { s.maxBytes = capBytes })
}

// checkByteCap refuses one source whose size exceeds the load's cap. Both
// file reads and in-memory loads call it, so a source cannot enter the
// decode pass unmeasured.
func checkByteCap(size int, file FilePath, limit byteCap) error {
	if byteCap(size) <= limit {
		return nil
	}

	return refuseAt(ErrSourceTooLarge, file, 1,
		gorefuse.Subject("the config file"),
		gorefuse.Why("split it into smaller files, or raise the size limit for this load."))
}
