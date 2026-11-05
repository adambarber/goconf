// The near miss: the same shapes, every type named.
package main

func main() {
	type (
		// Name is the named kind the strings become.
		Name string
		// Note is the named kind the field's note is.
		Note string
	)

	type record struct {
		note Note
	}

	names := []Name{""}

	r := record{note: Note("")}

	_, _ = names, r.note
}
