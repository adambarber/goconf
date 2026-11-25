package goconf

import (
	"fmt"
	"strings"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

// yamlDuplicateKeyTail is the yaml.v3 shape for a repeated mapping key.
const yamlDuplicateKeyTail = " already defined at line "

// yamlDuplicateKeyOf extracts the repeated key and its first line from one
// TypeError, before generic type classification can mislabel it.
func yamlDuplicateKeyOf(typeErr *yaml.TypeError) (ObjectKey, int, bool) {
	for _, e := range typeErr.Errors {
		head, tail, found := strings.Cut(e, yamlDuplicateKeyTail)
		if !found {
			continue
		}

		_, quoted, ok := strings.Cut(head, "mapping key ")
		if !ok {
			continue
		}

		var first int
		if _, scanErr := fmt.Sscanf(tail, "%d", &first); scanErr != nil || first <= 0 {
			continue
		}

		return ObjectKey(strings.Trim(quoted, `"`)), first, true
	}

	return "", 0, false
}

// yamlDuplicateRefusal classifies a repeated mapping key before generic
// type classification can mislabel it. Both the decode and presence walks
// call it, so the two boundaries cannot disagree on the reason.
func yamlDuplicateRefusal(file FilePath, typeErr *yaml.TypeError) error {
	key, first, found := yamlDuplicateKeyOf(typeErr)
	if !found {
		return nil
	}

	return refuseAt(ErrDuplicateKey, file, gorefuse.Line(first),
		gorefuse.Subject(`"`+string(key)+`"`),
		gorefuse.Why(`delete one of the two "`+string(key)+`" lines.`))
}
