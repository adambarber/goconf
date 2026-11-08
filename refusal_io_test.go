package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adambarber/gorefuse"
)

// ioCase is one IO-fault shape under test.
type ioCase struct {
	err  error
	name caseName
}

// ioFaultCases builds one refusal per IO fault site shape.
func ioFaultCases() []ioCase {
	return []ioCase{
		{
			name: "directory listing",
			err:  refuseIO(gorefuse.At(gorefuse.File("d"), 0, gorefuse.Subject("could not read the directory"), gorefuse.Why("check the directory's permissions (it must be readable by your user).")), os.ErrPermission),
		},
		{
			name: "candidate stat",
			err:  refuseIO(gorefuse.At(gorefuse.File("f"), 0, gorefuse.Subject("could not read the file"), gorefuse.Why("check the file's permissions (it must be readable by your user).")), os.ErrPermission),
		},
		{
			name: "file read",
			err:  ioReadRefusal("f", os.ErrPermission),
		},
		{
			name: "symlink resolution",
			err:  refuseIO(gorefuse.At(gorefuse.File("f"), 0, gorefuse.Subject("could not read the file"), gorefuse.Why("check that the file and its link target exist and are readable.")), os.ErrPermission),
		},
	}
}

// assertReasonIO fails unless err refuses with ErrIO.
func assertReasonIO(t *testing.T, err error, context caseName) {
	t.Helper()

	because, ok := errors.AsType[gorefuse.Because[Reason]](err)
	if !ok {
		t.Fatalf("%s: refusal %v must be a gorefuse.Because[Reason]", context, err)
	}

	if because.Reason != ErrIO {
		t.Fatalf("%s: reason = %v, want ErrIO", context, because.Reason)
	}
}

// TestIOFaultsCarryReasonIO pins that the IO-fault refusal carrier
// refuseIO builds — the shape every discovery and read fault site uses —
// refuses with ErrIO, not ErrDecoderRefused.
func TestIOFaultsCarryReasonIO(t *testing.T) {
	for _, tc := range ioFaultCases() {
		t.Run(string(tc.name), func(t *testing.T) {
			assertReasonIO(t, tc.err, tc.name)
		})
	}
}

// TestIOFaultsThroughLoadCarryReasonIO pins the discovery-stage IO faults
// end to end: a search path that cannot be listed refuses with ErrIO
// through Load, not ErrDecoderRefused.
func TestIOFaultsThroughLoadCarryReasonIO(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	t.Chdir(string(root))

	unlistable := filepath.Join(string(root), "unlistable")
	if err := os.MkdirAll(unlistable, 0); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.Chmod(unlistable, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	defer func() {
		_ = os.Remove(unlistable)
	}()

	_, err := discover(Discovery{Search: []DirPath{DirPath(unlistable)}, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatalf("unlistable directory must be refused")
	}

	assertReasonIO(t, err, caseName("through Load"))
}
