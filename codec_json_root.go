package goconf

import (
	"strings"

	"github.com/adambarber/gorefuse"
)

// jsonRootShapeRefusal refuses a file whose document root is not name:
// value pairs — the same fault YAML's not-a-mapping refusal names, with
// the same value vocabulary, whichever decoder noticed it.
func jsonRootShapeRefusal(file FilePath, line int, found goTypeName) error {
	word, _, _ := strings.Cut(string(found), " ")

	return refuseAt(ErrNotAMapping, file, gorefuse.Line(line),
		gorefuse.Subject("the file holds "+string(kindUserTerm(jsonWordKind(kindWord(word))))+
			", but settings must be name: value pairs"),
		gorefuse.Why(`start each setting on its own line as "name: value".`))
}
